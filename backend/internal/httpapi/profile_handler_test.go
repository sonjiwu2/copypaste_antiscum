package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProfileResetClearsCookieAndWeeklyTest(t *testing.T) {
	browser := newBrowser(NewRouter(testRouterDeps(t)))

	first := httptest.NewRecorder()
	browser.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/api/v1/weekly-tests/current", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("создать первый тест: %d %s", first.Code, first.Body.String())
	}
	oldProfile := browser.profileID()

	reset := httptest.NewRecorder()
	browser.ServeHTTP(reset, httptest.NewRequest(http.MethodDelete, "/api/v1/profile", nil))
	if reset.Code != http.StatusOK {
		t.Fatalf("сбросить профиль: %d %s", reset.Code, reset.Body.String())
	}
	if _, exists := browser.cookies[testCookieName]; exists {
		t.Fatal("после сброса browser сохранил cookie старого профиля")
	}

	second := httptest.NewRecorder()
	browser.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/api/v1/weekly-tests/current", nil))
	if second.Code != http.StatusOK {
		t.Fatalf("создать тест нового профиля: %d %s", second.Code, second.Body.String())
	}
	if browser.profileID() == oldProfile {
		t.Fatal("после сброса не был выдан новый анонимный профиль")
	}
}

func TestProfileIdentityAndLeaderboard(t *testing.T) {
	browser := newBrowser(NewRouter(testRouterDeps(t)))
	update := httptest.NewRecorder()
	browser.ServeHTTP(update, httptest.NewRequest(http.MethodPut, "/api/v1/profile",
		strings.NewReader(`{"displayName":"Дмитрий","avatar":"leader-2"}`)))
	if update.Code != http.StatusOK {
		t.Fatalf("обновить профиль: %d %s", update.Code, update.Body.String())
	}
	var identity profileIdentityResponse
	if err := json.Unmarshal(update.Body.Bytes(), &identity); err != nil {
		t.Fatalf("декодировать профиль: %v", err)
	}
	if identity.DisplayName != "Дмитрий" || identity.Avatar != "leader-2" {
		t.Fatalf("неожиданный профиль: %+v", identity)
	}

	leaders := httptest.NewRecorder()
	browser.ServeHTTP(leaders, httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard", nil))
	if leaders.Code != http.StatusOK {
		t.Fatalf("прочитать рейтинг: %d %s", leaders.Code, leaders.Body.String())
	}
}
