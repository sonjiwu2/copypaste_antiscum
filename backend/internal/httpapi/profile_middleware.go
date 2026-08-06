package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

// CookieSettings описывает cookie анонимного профиля.
type CookieSettings struct {
	Name   string
	MaxAge int

	// Secure выключается только для локального демо по HTTP.
	// В публичном развёртывании обязателен HTTPS и Secure=true.
	Secure bool
}

// withProfile выдаёт и подтверждает анонимный профиль запроса.
//
// Cookie помечена HttpOnly: идентификатор профиля — секрет доступа к истории
// пользователя, и JavaScript странице он не нужен. SameSite=Lax оставляет
// cookie работоспособной при обычных переходах и закрывает межсайтовые
// запросы, меняющие состояние.
//
// Синтаксически невалидное значение заменяется новым. Валидная cookie является
// bearer-секретом, поэтому её энтропия и конфиденциальность защищают историю.
func withProfile(profiles *profile.Service, settings CookieSettings, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		presented := profile.ID("")

		if cookie, err := r.Cookie(settings.Name); err == nil {
			presented = profile.ID(cookie.Value)
		}

		resolved, err := profiles.Resolve(r.Context(), presented)
		if err != nil {
			loggerFrom(r.Context()).ErrorContext(r.Context(), "не удалось определить профиль",
				slog.String("error", err.Error()))

			writeError(w, r, http.StatusInternalServerError, CodeInternalError,
				"Внутренняя ошибка сервера.")

			return
		}

		// Cookie переустанавливается только при выдаче нового профиля:
		// иначе каждый ответ менял бы состояние браузера без причины.
		if resolved.Issued {
			http.SetCookie(w, &http.Cookie{
				Name:     settings.Name,
				Value:    string(resolved.ID),
				Path:     "/",
				MaxAge:   settings.MaxAge,
				HttpOnly: true,
				Secure:   settings.Secure,
				SameSite: http.SameSiteLaxMode,
			})
		}

		next.ServeHTTP(w, r.WithContext(withProfileID(r.Context(), resolved.ID)))
	})
}
