package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// profileCookie возвращает cookie профиля из ответа.
func profileCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == testCookieName {
			return cookie
		}
	}

	return nil
}

func TestProfileCookieIsIssuedOnFirstRequest(t *testing.T) {
	router := NewRouter(testRouterDeps(t))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil))

	cookie := profileCookie(t, recorder)
	if cookie == nil {
		t.Fatal("cookie профиля не выдана")
	}

	if cookie.Value == "" {
		t.Error("cookie профиля пуста")
	}

	// HttpOnly закрывает идентификатор от скриптов страницы: он является
	// секретом доступа к истории пользователя.
	if !cookie.HttpOnly {
		t.Error("cookie профиля должна быть HttpOnly")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, ожидался Lax", cookie.SameSite)
	}

	if cookie.Path != "/" {
		t.Errorf("Path = %q, ожидался /", cookie.Path)
	}

	if cookie.MaxAge <= 0 {
		t.Errorf("Max-Age = %d, профиль должен жить долго", cookie.MaxAge)
	}
}

// Повторная выдача cookie на каждый ответ означала бы, что профиль
// пересоздаётся, а вместе с ним теряется история.
func TestProfileCookieIsNotReissued(t *testing.T) {
	router := NewRouter(testRouterDeps(t))

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil))

	issued := profileCookie(t, first)
	if issued == nil {
		t.Fatal("cookie профиля не выдана")
	}

	second := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
	request.AddCookie(&http.Cookie{Name: testCookieName, Value: issued.Value})
	router.ServeHTTP(second, request)

	if profileCookie(t, second) != nil {
		t.Error("cookie переустанавливается при каждом запросе")
	}
}

// Клиент не может выбрать себе профиль: непохожее значение заменяется новым.
func TestProfileCookieRejectsForeignValue(t *testing.T) {
	router := NewRouter(testRouterDeps(t))

	testCases := []struct {
		name  string
		value string
	}{
		{name: "пустое значение", value: ""},
		{name: "слишком короткое", value: "abc"},
		{name: "недопустимые символы", value: "profile id with spaces!!"},
		{name: "слишком длинное", value: strings.Repeat("a", 200)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
			request.AddCookie(&http.Cookie{Name: testCookieName, Value: testCase.value})

			router.ServeHTTP(recorder, request)

			cookie := profileCookie(t, recorder)
			if cookie == nil {
				t.Fatal("сервер должен выдать собственный профиль")
			}

			if cookie.Value == testCase.value {
				t.Error("сервер принял присланный клиентом профиль")
			}
		})
	}
}

// Проверки состояния идут мимо профиля: healthcheck опрашивает их постоянно
// и не должен создавать профиль на каждый опрос.
func TestHealthEndpointsDoNotIssueProfile(t *testing.T) {
	router := NewRouter(testRouterDeps(t))

	for _, target := range []string{"/healthz", "/readyz"} {
		t.Run(target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

			if profileCookie(t, recorder) != nil {
				t.Errorf("%s выдаёт cookie профиля", target)
			}
		})
	}
}

// Попытка одного пользователя недоступна другому даже при известном
// идентификаторе: он больше не является единственной защитой.
func TestAttemptsAreIsolatedBetweenProfiles(t *testing.T) {
	router := NewRouter(testRouterDeps(t))

	owner := newBrowser(router)
	stranger := newBrowser(router)

	started := startAttempt(t, owner, "buyer-fake-delivery")

	recorder := httptest.NewRecorder()
	stranger.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet,
		"/api/v1/attempts/"+started.AttemptID, nil))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("статус = %d, ожидался 403, тело: %s", recorder.Code, recorder.Body.String())
	}

	assertErrorCode(t, recorder.Body.Bytes(), CodeAttemptForbidden)

	// Чужой выбор тоже отклоняется.
	conflict := submitChoiceRaw(t, stranger, started.AttemptID,
		`{"nodeId":"`+started.CurrentNodeID+`","choiceId":"stay-on-platform","idempotencyKey":"key-1"}`)

	if conflict.Code != http.StatusForbidden {
		t.Errorf("статус = %d, ожидался 403", conflict.Code)
	}

	// Владелец продолжает работать со своей попыткой.
	if owner.profileID() == stranger.profileID() {
		t.Fatal("профили клиентов совпали")
	}

	submitChoiceOK(t, owner, started.AttemptID, started.CurrentNodeID, "stay-on-platform", "key-1")
}
