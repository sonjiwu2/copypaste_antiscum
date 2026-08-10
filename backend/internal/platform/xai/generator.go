// Package xai обращается к API xAI и преобразует строгий JSON-ответ Grok
// в доменный еженедельный тест.
package xai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type Generator struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewGenerator(config Config) *Generator {
	return &Generator{
		apiKey:  config.APIKey,
		baseURL: strings.TrimRight(config.BaseURL, "/"),
		model:   config.Model,
		client:  &http.Client{Timeout: config.Timeout},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model               string        `json:"model"`
	Messages            []chatMessage `json:"messages"`
	Temperature         float64       `json:"temperature"`
	MaxCompletionTokens int           `json:"max_completion_tokens"`
	ReasoningEffort     string        `json:"reasoning_effort"`
	ResponseFormat      any           `json:"response_format"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type generatedPayload struct {
	Title     string `json:"title"`
	Intro     string `json:"intro"`
	Questions []struct {
		Prompt       string                `json:"prompt"`
		Options      []string              `json:"options"`
		CorrectIndex int                   `json:"correctIndex"`
		Explanation  string                `json:"explanation"`
		RiskTag      string                `json:"riskTag"`
		Difficulty   weeklytest.Difficulty `json:"difficulty"`
	} `json:"questions"`
}

func (g *Generator) Generate(
	ctx context.Context,
	_ weeklytest.GenerationContext,
) (weeklytest.Generated, error) {
	if g.apiKey == "" {
		return weeklytest.Generated{}, fmt.Errorf("XAI_API_KEY не задан")
	}

	payload := chatRequest{
		Model: g.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Создай новый еженедельный тест на эту неделю."},
		},
		Temperature:         0.35,
		MaxCompletionTokens: 5600,
		ReasoningEffort:     "none",
		ResponseFormat:      weeklyTestResponseFormat(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return weeklytest.Generated{}, fmt.Errorf("сериализовать запрос Grok: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body),
	)
	if err != nil {
		return weeklytest.Generated{}, fmt.Errorf("создать запрос Grok: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+g.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := g.client.Do(request)
	if err != nil {
		return weeklytest.Generated{}, fmt.Errorf("выполнить запрос Grok: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// Тело ошибки не логируется: сторонний сервис может отразить фрагменты
		// запроса. Для диагностики достаточно статуса и request-id.
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024))
		return weeklytest.Generated{}, fmt.Errorf(
			"Grok ответил HTTP %d (request-id: %s)",
			response.StatusCode,
			response.Header.Get("x-request-id"),
		)
	}

	limited := io.LimitReader(response.Body, 1024*1024)
	var decoded chatResponse
	if err := json.NewDecoder(limited).Decode(&decoded); err != nil {
		return weeklytest.Generated{}, fmt.Errorf("разобрать ответ Grok: %w", err)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return weeklytest.Generated{}, fmt.Errorf("Grok вернул пустой ответ")
	}

	var generated generatedPayload
	if err := json.Unmarshal([]byte(decoded.Choices[0].Message.Content), &generated); err != nil {
		return weeklytest.Generated{}, fmt.Errorf("разобрать JSON теста Grok: %w", err)
	}

	result := weeklytest.Generated{
		Title:     generated.Title,
		Intro:     generated.Intro,
		Source:    weeklytest.SourceGrok,
		Model:     g.model,
		Questions: make([]weeklytest.Question, 0, len(generated.Questions)),
	}
	for _, question := range generated.Questions {
		result.Questions = append(result.Questions, weeklytest.Question{
			Prompt:       question.Prompt,
			Options:      question.Options,
			CorrectIndex: question.CorrectIndex,
			Explanation:  question.Explanation,
			RiskTag:      question.RiskTag,
			Difficulty:   question.Difficulty,
		})
	}

	return result, nil
}

const systemPrompt = `Ты создаёшь учебный тест на русском языке для антискам-тренажёра онлайн-сделок.
Составь ровно 20 оригинальных практических вопросов. У каждого ровно 4 правдоподобных варианта и только один безопасный правильный ответ.
Темы: фишинговые ссылки и QR-коды, коды из SMS, предоплата, поддельная поддержка, срочность, безопасная сделка, удалённый доступ, социальная инженерия.
Объяснение должно кратко раскрывать конкретный признак риска и безопасное действие. Не проси пользователя реально переходить по ссылкам, звонить, платить или раскрывать данные.
Все поля и варианты ответа пиши по-русски; riskTag оставляй коротким машинным кодом на английском.`

func weeklyTestResponseFormat() map[string]any {
	questionSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"prompt", "options", "correctIndex", "explanation", "riskTag", "difficulty",
		},
		"properties": map[string]any{
			"prompt": map[string]any{"type": "string", "minLength": 12, "maxLength": 500},
			"options": map[string]any{
				"type": "array", "minItems": 4, "maxItems": 4,
				"items": map[string]any{"type": "string", "minLength": 1, "maxLength": 240},
			},
			"correctIndex": map[string]any{"type": "integer", "minimum": 0, "maximum": 3},
			"explanation":  map[string]any{"type": "string", "minLength": 12, "maxLength": 700},
			"riskTag":      map[string]any{"type": "string", "minLength": 2, "maxLength": 60},
			"difficulty": map[string]any{
				"type": "string", "enum": []string{"easy", "medium", "hard"},
			},
		},
	}

	return map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   "weekly_antiscam_test",
			"strict": true,
			"schema": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"title", "intro", "questions"},
				"properties": map[string]any{
					"title": map[string]any{"type": "string", "minLength": 3, "maxLength": 100},
					"intro": map[string]any{"type": "string", "minLength": 10, "maxLength": 300},
					"questions": map[string]any{
						"type": "array", "minItems": 20, "maxItems": 20, "items": questionSchema,
					},
				},
			},
		},
	}
}
