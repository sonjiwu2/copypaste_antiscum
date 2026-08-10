package httpapi

import (
	"net/http"
	"slices"
	"strconv"
	"time"
)

// preflightMaxAge — насколько долго браузер может кэшировать разрешение.
// Значение умеренное: смена списка origin не должна ждать сутки.
const preflightMaxAge = 10 * time.Minute

// allowedRequestHeaders — заголовки, которые фронтенд имеет право прислать.
// Cookie в списке нет намеренно: её отправку разрешает Allow-Credentials,
// а не перечисление заголовков.
const allowedRequestHeaders = "Content-Type, X-Request-ID"

// allowedMethods — методы публичного API.
const allowedMethods = "GET, POST, DELETE, OPTIONS"

// CORSSettings описывает межсайтовый доступ браузерного клиента.
type CORSSettings struct {
	// AllowedOrigins перечисляет origin фронтенда целиком (scheme://host[:port]).
	// Пустой список означает, что кросс-доменный доступ выключен.
	AllowedOrigins []string
}

// withCORS разрешает браузерному клиенту с известного origin ходить в API
// вместе с cookie профиля.
//
// Ответ всегда называет конкретный origin, а не "*": ответы API привязаны к
// анонимному профилю, и маска с credentials браузером запрещена и опасна.
// Запрос с неизвестного origin проходит дальше без CORS-заголовков — его
// заблокирует сам браузер, а серверный клиент (curl, тесты, мониторинг)
// продолжает работать как раньше.
func withCORS(settings CORSSettings, next http.Handler) http.Handler {
	if len(settings.AllowedOrigins) == 0 {
		return next
	}

	allowed := slices.Clone(settings.AllowedOrigins)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Ответ зависит от Origin даже когда заголовок отсутствует: без
		// Vary общий кэш отдал бы чужому origin разрешающий ответ.
		w.Header().Add("Vary", "Origin")

		if origin == "" || !slices.Contains(allowed, origin) {
			// Preflight с неизвестного origin дальше не идёт: маршруты
			// OPTIONS не обслуживают, а профиль создавать незачем.
			if isPreflight(r) {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)

			return
		}

		header := w.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Set("Access-Control-Expose-Headers", requestHeaderName)

		if isPreflight(r) {
			header.Add("Vary", "Access-Control-Request-Method")
			header.Add("Vary", "Access-Control-Request-Headers")
			header.Set("Access-Control-Allow-Methods", allowedMethods)
			header.Set("Access-Control-Allow-Headers", allowedRequestHeaders)
			header.Set("Access-Control-Max-Age", strconv.Itoa(int(preflightMaxAge.Seconds())))

			// Предварительный запрос не доходит до профиля и обработчиков:
			// иначе каждый OPTIONS создавал бы анонимный профиль.
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// isPreflight отличает предварительный запрос от обычного OPTIONS.
func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
