package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const frontendOrigin = "http://localhost:5173"

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	router := newRouterWithCORS(t, CORSSettings{AllowedOrigins: []string{frontendOrigin}})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
	request.Header.Set("Origin", frontendOrigin)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался %d", recorder.Code, http.StatusOK)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != frontendOrigin {
		t.Fatalf("Access-Control-Allow-Origin = %q, ожидался %q", got, frontendOrigin)
	}

	// Без Allow-Credentials браузер не отправит cookie профиля, и вся история
	// пользователя станет недоступной.
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, ожидался %q", got, "true")
	}

	if !strings.Contains(recorder.Header().Get("Vary"), "Origin") {
		t.Fatalf("Vary = %q, ожидался с Origin", recorder.Header().Get("Vary"))
	}
}

func TestCORSIgnoresUnknownOrigin(t *testing.T) {
	t.Parallel()

	router := newRouterWithCORS(t, CORSSettings{AllowedOrigins: []string{frontendOrigin}})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
	request.Header.Set("Origin", "http://evil.example")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, ожидался пустым", got)
	}
}

func TestCORSPreflightDoesNotIssueProfile(t *testing.T) {
	t.Parallel()

	router := newRouterWithCORS(t, CORSSettings{AllowedOrigins: []string{frontendOrigin}})

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/attempts", nil)
	request.Header.Set("Origin", frontendOrigin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("статус = %d, ожидался %d", recorder.Code, http.StatusNoContent)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPost) {
		t.Fatalf("Access-Control-Allow-Methods = %q, ожидался с POST", got)
	}

	// Предварительный запрос не должен доходить до профиля: иначе браузер
	// заводил бы анонимный профиль на каждый preflight.
	if cookies := recorder.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("preflight выдал cookie: %v", cookies)
	}
}

func TestCORSPreflightFromUnknownOriginIsNotAllowed(t *testing.T) {
	t.Parallel()

	router := newRouterWithCORS(t, CORSSettings{AllowedOrigins: []string{frontendOrigin}})

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/attempts", nil)
	request.Header.Set("Origin", "http://evil.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("статус = %d, ожидался %d", recorder.Code, http.StatusNoContent)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, ожидался пустым", got)
	}

	if cookies := recorder.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("preflight выдал cookie: %v", cookies)
	}
}

func TestCORSDisabledByDefault(t *testing.T) {
	t.Parallel()

	router := newRouterWithCORS(t, CORSSettings{})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
	request.Header.Set("Origin", frontendOrigin)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался %d", recorder.Code, http.StatusOK)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, ожидался пустым", got)
	}

	// Серверный клиент без браузера продолжает работать как раньше.
	if got := recorder.Header().Get("Vary"); strings.Contains(got, "Origin") {
		t.Fatalf("Vary = %q, ожидался без Origin при выключенном CORS", got)
	}
}

func TestCookieSameSiteIsConfigurable(t *testing.T) {
	t.Parallel()

	deps := testRouterDeps(t)
	deps.Cookie.Secure = true
	deps.Cookie.SameSite = http.SameSiteNoneMode
	router := NewRouter(deps)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil))

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("выдано cookie: %d, ожидалась 1", len(cookies))
	}

	if cookies[0].SameSite != http.SameSiteNoneMode {
		t.Fatalf("SameSite = %v, ожидался %v", cookies[0].SameSite, http.SameSiteNoneMode)
	}

	if !cookies[0].Secure {
		t.Fatal("cookie с SameSite=None обязана быть Secure")
	}
}

// newRouterWithCORS собирает роутер с заданной политикой межсайтового доступа.
func newRouterWithCORS(t *testing.T, settings CORSSettings) http.Handler {
	t.Helper()

	deps := testRouterDeps(t)
	deps.CORS = settings

	return NewRouter(deps)
}
