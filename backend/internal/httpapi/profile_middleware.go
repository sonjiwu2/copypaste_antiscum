package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/auth"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

// CookieSettings описывает cookie анонимного профиля.
type CookieSettings struct {
	Name   string
	MaxAge int

	// Secure выключается только для локального демо по HTTP.
	// В публичном развёртывании обязателен HTTPS и Secure=true.
	Secure bool

	// SameSite задаёт поведение при межсайтовых запросах. Нулевое значение
	// означает Lax: фронтенд обслуживается тем же сайтом, что и API.
	SameSite http.SameSite
}

// sameSite подставляет Lax вместо незаданной политики.
//
// Нулевое значение http.SameSite и SameSiteDefaultMode отправили бы cookie
// без атрибута; браузеры трактуют такую cookie как Lax, но писать политику в
// заголовке явно надёжнее.
func (s CookieSettings) sameSite() http.SameSite {
	switch s.SameSite {
	case http.SameSiteLaxMode, http.SameSiteStrictMode, http.SameSiteNoneMode:
		return s.SameSite
	default:
		return http.SameSiteLaxMode
	}
}

// withProfile выдаёт и подтверждает анонимный профиль запроса.
//
// Cookie помечена HttpOnly: идентификатор профиля — секрет доступа к истории
// пользователя, и JavaScript странице он не нужен. SameSite по умолчанию Lax:
// обычные переходы работают, а межсайтовые запросы, меняющие состояние, — нет.
// Кросс-доменному фронтенду нужен SameSite=None вместе с HTTPS и списком
// разрешённых origin.
//
// Синтаксически невалидное значение заменяется новым. Валидная cookie является
// bearer-секретом, поэтому её энтропия и конфиденциальность защищают историю.
func withPrincipal(profiles *profile.Service, authentication *auth.Service,
	profileCookie, sessionCookie CookieSettings, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(sessionCookie.Name); err == nil && cookie.Value != "" {
			current, resolveErr := authentication.Resolve(r.Context(), cookie.Value)
			if resolveErr == nil {
				ctx := withProfileID(r.Context(), current.ProfileID)
				next.ServeHTTP(w, r.WithContext(withAccount(ctx, current)))
				return
			}
			if resolveErr != auth.ErrRequired {
				loggerFrom(r.Context()).ErrorContext(r.Context(), "не удалось прочитать сессию",
					slog.String("error", resolveErr.Error()))
				writeError(w, r, http.StatusInternalServerError, CodeInternalError,
					"Внутренняя ошибка сервера.")
				return
			}
			clearCookie(w, sessionCookie)
		}

		// Вход, выход и проверка сессии не нуждаются во временном игровом
		// профиле. Регистрация — исключение: она привязывает существующий
		// анонимный прогресс либо только что выданный профиль к аккаунту.
		if r.URL.Path != "/api/v1/auth/register" &&
			(r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/logout" ||
				r.URL.Path == "/api/v1/auth/session") {
			next.ServeHTTP(w, r)
			return
		}

		presented := profile.ID("")

		if cookie, err := r.Cookie(profileCookie.Name); err == nil {
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
				Name:     profileCookie.Name,
				Value:    string(resolved.ID),
				Path:     "/",
				MaxAge:   profileCookie.MaxAge,
				HttpOnly: true,
				Secure:   profileCookie.Secure,
				SameSite: profileCookie.sameSite(),
			})
		}

		next.ServeHTTP(w, r.WithContext(withProfileID(r.Context(), resolved.ID)))
	})
}

func requireAccount(allowAnonymous bool, next http.Handler) http.Handler {
	if allowAnonymous {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := AccountFrom(r.Context()); !ok {
			writeError(w, r, http.StatusUnauthorized, CodeAuthRequired, "Войдите в аккаунт.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setSessionCookie(w http.ResponseWriter, settings CookieSettings, token string) {
	http.SetCookie(w, &http.Cookie{Name: settings.Name, Value: token, Path: "/",
		MaxAge: settings.MaxAge, HttpOnly: true, Secure: settings.Secure,
		SameSite: settings.sameSite()})
}

func clearCookie(w http.ResponseWriter, settings CookieSettings) {
	http.SetCookie(w, &http.Cookie{Name: settings.Name, Value: "", Path: "/", MaxAge: -1,
		Expires: time.Unix(1, 0).UTC(), HttpOnly: true, Secure: settings.Secure,
		SameSite: settings.sameSite()})
}
