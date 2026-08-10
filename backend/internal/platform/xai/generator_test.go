package xai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/xai"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

func TestGeneratorCallsXAIWithServerKeyAndStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" || r.Method != http.MethodPost {
			t.Fatalf("неожиданный запрос: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-xai-key" {
			t.Fatalf("Authorization = %q", got)
		}

		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("разобрать запрос: %v", err)
		}
		format, ok := request["response_format"].(map[string]any)
		if !ok || format["type"] != "json_schema" {
			t.Fatalf("не задан structured output: %#v", request["response_format"])
		}

		questions := make([]map[string]any, weeklytest.QuestionCount)
		for index := range questions {
			questions[index] = map[string]any{
				"prompt": "Практический вопрос цифровой безопасности?", "options": []string{"A", "B", "C", "D"},
				"correctIndex": index % 4, "explanation": "Подробное безопасное объяснение.",
				"riskTag": "phishing", "difficulty": "medium",
			}
		}
		contentJSON, err := json.Marshal(map[string]any{
			"title": "Новый тест", "intro": "Выберите самое безопасное действие в каждой ситуации.",
			"questions": questions,
		})
		if err != nil {
			t.Fatalf("собрать ответ: %v", err)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": string(contentJSON)}}},
		})
	}))
	defer server.Close()

	generator := xai.NewGenerator(xai.Config{
		APIKey: "secret-xai-key", BaseURL: server.URL, Model: "grok-test", Timeout: time.Second,
	})
	generated, err := generator.Generate(context.Background(), weeklytest.GenerationContext{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if generated.Source != weeklytest.SourceGrok || generated.Model != "grok-test" || len(generated.Questions) != weeklytest.QuestionCount {
		t.Fatalf("неожиданный результат: %+v", generated)
	}
}
