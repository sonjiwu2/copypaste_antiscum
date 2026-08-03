package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
)

// RouterDeps — зависимости HTTP-слоя.
type RouterDeps struct {
	Logger          *slog.Logger
	RequestIDs      identifier.Generator
	MaxRequestBytes int64
}

// NewRouter собирает маршруты и цепочку middleware.
func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealth)
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
