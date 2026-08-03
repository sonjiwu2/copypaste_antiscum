package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// RouterDeps — зависимости HTTP-слоя.
type RouterDeps struct {
	Logger          *slog.Logger
	RequestIDs      identifier.Generator
	MaxRequestBytes int64
	Scenarios       *scenario.Service
	Attempts        *attempt.Service
}

// NewRouter собирает маршруты и цепочку middleware.
func NewRouter(deps RouterDeps) http.Handler {
	scenarios := &scenarioHandler{scenarios: deps.Scenarios}
	attempts := &attemptHandler{attempts: deps.Attempts}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /api/v1/scenarios", scenarios.list)
	mux.HandleFunc("GET /api/v1/scenarios/{scenarioId}", scenarios.get)
	mux.HandleFunc("POST /api/v1/attempts", attempts.start)
	mux.HandleFunc("GET /api/v1/attempts/{attemptId}", attempts.get)
	mux.HandleFunc("POST /api/v1/attempts/{attemptId}/choices", attempts.submitChoice)
	// Общий маршрут перехватывает неизвестные пути, чтобы клиент всегда
	// получал JSON-ошибку вместо стандартного текстового ответа ServeMux.
	mux.HandleFunc("/", handleNotFound)

	var handler http.Handler = mux
	handler = withBodyLimit(deps.MaxRequestBytes, handler)
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

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusNotFound, CodeNotFound, "Запрошенный ресурс не найден.")
}
