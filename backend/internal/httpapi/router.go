package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/auth"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

// ReadinessCheck подтверждает, что приложение способно обслуживать запросы.
//
// Проверка приходит из сборки приложения: HTTP-слой не должен знать,
// какие именно внешние системы нужны сервису.
type ReadinessCheck func(ctx context.Context) error

// RouterDeps — зависимости HTTP-слоя.
type RouterDeps struct {
	Logger          *slog.Logger
	RequestIDs      identifier.Generator
	MaxRequestBytes int64
	Ready           ReadinessCheck
	Cookie          CookieSettings
	SessionCookie   CookieSettings
	CORS            CORSSettings
	AllowAnonymous  bool
	Authentication  *auth.Service
	Profiles        *profile.Service
	Scenarios       *scenario.Service
	Attempts        *attempt.Service
	Progress        *progress.Service
	WeeklyTests     *weeklytest.Service
}

// NewRouter собирает маршруты и цепочку middleware.
func NewRouter(deps RouterDeps) http.Handler {
	scenarios := &scenarioHandler{scenarios: deps.Scenarios}
	attempts := &attemptHandler{attempts: deps.Attempts}
	readiness := &readinessHandler{check: deps.Ready}
	userProgress := &progressHandler{progress: deps.Progress}
	weeklyTests := &weeklyTestHandler{tests: deps.WeeklyTests}
	profiles := &profileHandler{profiles: deps.Profiles, cookie: deps.Cookie, sessionCookie: deps.SessionCookie}
	authentication := &authHandler{auth: deps.Authentication, cookie: deps.SessionCookie}

	protected := http.NewServeMux()

	protected.HandleFunc("GET /api/v1/scenarios", scenarios.list)
	protected.HandleFunc("GET /api/v1/scenarios/{scenarioId}", scenarios.get)
	protected.HandleFunc("POST /api/v1/attempts", attempts.start)
	protected.HandleFunc("GET /api/v1/attempts/{attemptId}", attempts.get)
	protected.HandleFunc("POST /api/v1/attempts/{attemptId}/choices", attempts.submitChoice)
	protected.HandleFunc("GET /api/v1/progress", userProgress.get)
	protected.HandleFunc("GET /api/v1/leaderboard", profiles.leaderboard)
	protected.HandleFunc("PUT /api/v1/profile", profiles.update)
	protected.HandleFunc("DELETE /api/v1/profile", profiles.reset)
	protected.HandleFunc("POST /api/v1/weekly-tests/current", weeklyTests.current)
	protected.HandleFunc("POST /api/v1/weekly-tests/{testId}/answer", weeklyTests.checkAnswer)
	protected.HandleFunc("POST /api/v1/weekly-tests/{testId}/submit", weeklyTests.submit)
	// Общий маршрут перехватывает неизвестные пути, чтобы клиент всегда
	// получал JSON-ошибку вместо стандартного текстового ответа ServeMux.
	protected.HandleFunc("/", handleNotFound)

	api := http.NewServeMux()
	api.HandleFunc("POST /api/v1/auth/register", authentication.register)
	api.HandleFunc("POST /api/v1/auth/login", authentication.login)
	api.HandleFunc("POST /api/v1/auth/logout", authentication.logout)
	api.HandleFunc("GET /api/v1/auth/session", authentication.current)
	api.Handle("/api/v1/", requireAccount(deps.AllowAnonymous, protected))

	mux := http.NewServeMux()

	// Проверки состояния идут мимо профиля: healthcheck контейнера ходит
	// постоянно и не должен создавать по профилю на каждый опрос.
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /readyz", readiness.get)
	mux.Handle("/api/v1/", withPrincipal(deps.Profiles, deps.Authentication,
		deps.Cookie, deps.SessionCookie, api))
	mux.HandleFunc("/", handleNotFound)

	var handler http.Handler = mux
	handler = withBodyLimit(deps.MaxRequestBytes, handler)
	// CORS стоит выше ограничения тела: предварительный запрос тела не имеет
	// и обязан получить ответ раньше маршрутизации и выдачи профиля.
	handler = withCORS(deps.CORS, handler)
	handler = withPanicRecovery(handler)
	handler = withAccessLog(handler)
	handler = withRequestContext(deps.Logger, deps.RequestIDs, handler)

	return handler
}

// healthResponse — тело ответа проверки живости.
type healthResponse struct {
	Status string `json:"status"`
}

// handleHealth намеренно не проверяет внешние системы: он подтверждает только
// то, что процесс жив и способен отвечать.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, healthResponse{Status: "ok"})
}

// readinessHandler отвечает на проверку готовности к обслуживанию запросов.
type readinessHandler struct {
	check ReadinessCheck
}

// get подтверждает готовность приложения.
//
// В отличие от /healthz проверка затрагивает внешние зависимости, поэтому
// балансировщик и Compose ориентируются именно на неё. Причина отказа уходит
// в лог, а не клиенту: устройство сервиса наружу не публикуется.
func (h *readinessHandler) get(w http.ResponseWriter, r *http.Request) {
	if h.check == nil {
		writeJSON(w, r, http.StatusOK, healthResponse{Status: "ready"})

		return
	}

	if err := h.check(r.Context()); err != nil {
		loggerFrom(r.Context()).ErrorContext(r.Context(), "приложение не готово обслуживать запросы",
			slog.String("error", err.Error()))

		writeJSON(w, r, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})

		return
	}

	writeJSON(w, r, http.StatusOK, healthResponse{Status: "ready"})
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusNotFound, CodeNotFound, "Запрошенный ресурс не найден.")
}
