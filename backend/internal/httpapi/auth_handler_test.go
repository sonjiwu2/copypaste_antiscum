package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterLoginSessionAndLogout(t *testing.T) {
	router := NewRouter(testRouterDeps(t))
	firstDevice := newBrowser(router)
	secondDevice := newBrowser(router)

	registered := httptest.NewRecorder()
	firstDevice.ServeHTTP(registered, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(`{"email":"Dmitriy@example.com","password":"password1!","displayName":"Дмитрий","avatar":"leader-2"}`)))
	if registered.Code != http.StatusCreated {
		t.Fatalf("регистрация: %d %s", registered.Code, registered.Body.String())
	}
	assertSessionResponse(t, registered.Body.Bytes(), "Dmitriy@example.com", "Дмитрий")

	duplicate := httptest.NewRecorder()
	secondDevice.ServeHTTP(duplicate, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(`{"email":"dmitriy@example.com","password":"password2!","displayName":"Другой","avatar":"profile"}`)))
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("дубликат почты: %d %s", duplicate.Code, duplicate.Body.String())
	}
	assertErrorCode(t, duplicate.Body.Bytes(), CodeEmailTaken)

	wrong := httptest.NewRecorder()
	secondDevice.ServeHTTP(wrong, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"Dmitriy@example.com","password":"wrongpass!"}`)))
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("неверный пароль: %d", wrong.Code)
	}

	loggedIn := httptest.NewRecorder()
	secondDevice.ServeHTTP(loggedIn, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"dmitriy@example.com","password":"password1!"}`)))
	if loggedIn.Code != http.StatusOK {
		t.Fatalf("вход: %d %s", loggedIn.Code, loggedIn.Body.String())
	}
	assertSessionResponse(t, loggedIn.Body.Bytes(), "Dmitriy@example.com", "Дмитрий")

	current := httptest.NewRecorder()
	secondDevice.ServeHTTP(current, httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil))
	if current.Code != http.StatusOK {
		t.Fatalf("восстановление сессии: %d %s", current.Code, current.Body.String())
	}

	loggedOut := httptest.NewRecorder()
	secondDevice.ServeHTTP(loggedOut, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	if loggedOut.Code != http.StatusOK {
		t.Fatalf("выход: %d", loggedOut.Code)
	}
	afterLogout := httptest.NewRecorder()
	secondDevice.ServeHTTP(afterLogout, httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil))
	if afterLogout.Code != http.StatusUnauthorized {
		t.Fatalf("сессия после выхода: %d", afterLogout.Code)
	}
}

func assertSessionResponse(t *testing.T, body []byte, email, displayName string) {
	t.Helper()
	var response sessionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("session JSON: %v", err)
	}
	if response.Email != email || response.DisplayName != displayName || response.ExpiresAt == "" {
		t.Fatalf("неожиданная сессия: %+v", response)
	}
}
