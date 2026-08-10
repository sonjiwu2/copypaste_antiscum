package groq_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/groq"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

func TestGeneratorCallsGroqWithServerKeyAndStrictStructuredOutput(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/chat/completions" || r.Method != http.MethodPost {
			t.Fatalf("неожиданный запрос: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-groq-key" {
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
		jsonSchema, ok := format["json_schema"].(map[string]any)
		if !ok || jsonSchema["strict"] != true {
			t.Fatalf("не включён strict mode: %#v", format["json_schema"])
		}
		messages, ok := request["messages"].([]any)
		if !ok || len(messages) < 2 {
			t.Fatalf("не переданы инструкции: %#v", request["messages"])
		}
		systemMessage, ok := messages[0].(map[string]any)
		if !ok || !strings.Contains(systemMessage["content"].(string), "молча проведи аудит каждого вопроса") {
			t.Fatalf("в системной инструкции нет аудита качества")
		}

		// Реальные модели иногда возвращают один лишний вопрос. Генератор должен
		// принять валидный ответ и обрезать его до доменного количества.
		questions := make([]map[string]any, weeklytest.QuestionCount+1)
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

	generator := groq.NewGenerator(groq.Config{
		APIKey: "test-groq-key", BaseURL: server.URL, Model: "openai/gpt-oss-20b", Timeout: time.Second,
	})
	generated, err := generator.Generate(context.Background(), weeklytest.GenerationContext{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if generated.Source != weeklytest.SourceGroq || generated.Model != "openai/gpt-oss-20b" || len(generated.Questions) != weeklytest.QuestionCount {
		t.Fatalf("неожиданный результат: %+v", generated)
	}
	if requests != 1 {
		t.Fatalf("запросов к Groq = %d, ожидался 1", requests)
	}
}
