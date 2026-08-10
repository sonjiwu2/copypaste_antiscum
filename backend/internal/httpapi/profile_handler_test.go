package httpapi

import (
	"net/http"
	"net/http/httptest"
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
