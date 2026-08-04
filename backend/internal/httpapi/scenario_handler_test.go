package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// hiddenFields — поля, которые не должны появляться в ответах каталога.
// Их утечка означала бы, что клиент видит содержимое сценария заранее.
var hiddenFields = []string{
	"nodes", "node", "choices", "choice", "consequence", "outcome",
	"nextNodeId", "startNodeId", "safetyScore", "criticality",
	"riskTags", "skillEffects", "isActive", "text", "sender",
}

func doRequest(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

	return recorder
}

func TestListScenarios(t *testing.T) {
	testCases := []struct {
		name       string
		target     string
		wantStatus int
		wantIDs    []string
		wantCode   string
	}{
		{
			name:       "весь каталог",
			target:     "/api/v1/scenarios",
			wantStatus: http.StatusOK,
			wantIDs: []string{
				"buyer-fake-delivery",
				"buyer-iphone-deposit",
				"buyer-ps5-delivery",
				"seller-gpu-return-swap",
				"seller-laptop-courier",
				"seller-payment-already-sent",
			},
		},
		{
			name:       "фильтр по роли покупателя",
			target:     "/api/v1/scenarios?role=buyer",
			wantStatus: http.StatusOK,
			wantIDs: []string{
				"buyer-fake-delivery",
				"buyer-iphone-deposit",
				"buyer-ps5-delivery",
			},
		},
		{
			name:       "фильтр по роли продавца",
			target:     "/api/v1/scenarios?role=seller",
			wantStatus: http.StatusOK,
			wantIDs: []string{
				"seller-gpu-return-swap",
				"seller-laptop-courier",
				"seller-payment-already-sent",
			},
		},
		{
			name:       "неподдерживаемая роль",
			target:     "/api/v1/scenarios?role=courier",
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeUnsupportedRole,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := doRequest(t, testCase.target)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("статус = %d, ожидался %d, тело: %s",
					recorder.Code, testCase.wantStatus, recorder.Body.String())
			}

			if testCase.wantCode != "" {
				assertErrorCode(t, recorder.Body.Bytes(), testCase.wantCode)

				return
			}

			var body scenarioListResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("не удалось разобрать ответ: %v", err)
			}

			gotIDs := make([]string, len(body.Scenarios))
			for i, item := range body.Scenarios {
				gotIDs[i] = item.ID
			}

			if strings.Join(gotIDs, ",") != strings.Join(testCase.wantIDs, ",") {
				t.Errorf("сценарии = %v, ожидались %v", gotIDs, testCase.wantIDs)
			}
		})
	}
}

func TestListScenariosReturnsUsableMetadata(t *testing.T) {
	recorder := doRequest(t, "/api/v1/scenarios?role=buyer")

	var body scenarioListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	if len(body.Scenarios) == 0 {
		t.Fatal("ожидался хотя бы один сценарий покупателя")
	}

	found := body.Scenarios[0]

	if found.Role != "buyer" {
		t.Errorf("role = %q, ожидалось buyer", found.Role)
	}

	if found.Title == "" || found.Description == "" || found.Difficulty == "" {
		t.Errorf("метаданные неполные: %+v", found)
	}

	if found.Version < 1 {
		t.Errorf("version = %d, ожидалась положительная", found.Version)
	}

	if found.EstimatedMinutes <= 0 {
		t.Errorf("estimatedMinutes = %d, ожидалось положительное", found.EstimatedMinutes)
	}
}

func TestGetScenario(t *testing.T) {
	testCases := []struct {
		name       string
		target     string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "существующий сценарий",
			target:     "/api/v1/scenarios/buyer-fake-delivery",
			wantStatus: http.StatusOK,
		},
		{
			name:       "неизвестный сценарий",
			target:     "/api/v1/scenarios/unknown-scenario",
			wantStatus: http.StatusNotFound,
			wantCode:   CodeScenarioNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := doRequest(t, testCase.target)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("статус = %d, ожидался %d", recorder.Code, testCase.wantStatus)
			}

			if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
				t.Errorf("Content-Type = %q, ожидался JSON", recorder.Header().Get("Content-Type"))
			}

			if testCase.wantCode != "" {
				assertErrorCode(t, recorder.Body.Bytes(), testCase.wantCode)

				return
			}

			var body scenarioSummary
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("не удалось разобрать ответ: %v", err)
			}

			if body.ID != "buyer-fake-delivery" {
				t.Errorf("id = %q, ожидался buyer-fake-delivery", body.ID)
			}
		})
	}
}

// Каталог не должен раскрывать содержимое сценария: ни узлов, ни вариантов
// выбора, ни внутренних коэффициентов.
func TestCatalogResponsesHideScenarioGraph(t *testing.T) {
	targets := []string{
		"/api/v1/scenarios",
		"/api/v1/scenarios?role=seller",
		"/api/v1/scenarios/seller-payment-already-sent",
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			recorder := doRequest(t, target)

			if recorder.Code != http.StatusOK {
				t.Fatalf("статус = %d, ожидался 200", recorder.Code)
			}

			var payload any
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatalf("не удалось разобрать ответ: %v", err)
			}

			for _, field := range collectJSONKeys(payload) {
				for _, hidden := range hiddenFields {
					if field == hidden {
						t.Errorf("в ответе присутствует скрытое поле %q", field)
					}
				}
			}
		})
	}
}

// collectJSONKeys собирает все имена полей ответа на любой глубине.
func collectJSONKeys(payload any) []string {
	switch typed := payload.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key, value := range typed {
			keys = append(keys, key)
			keys = append(keys, collectJSONKeys(value)...)
		}

		return keys
	case []any:
		keys := make([]string, 0, len(typed))
		for _, value := range typed {
			keys = append(keys, collectJSONKeys(value)...)
		}

		return keys
	default:
		return nil
	}
}
