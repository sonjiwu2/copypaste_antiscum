// Package groq обращается к OpenAI-совместимому API Groq Cloud и
// преобразует строгий JSON-ответ модели в доменный еженедельный тест.
package groq

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
		return weeklytest.Generated{}, fmt.Errorf("GROQ_API_KEY не задан")
	}

	result := weeklytest.Generated{
		Source: weeklytest.SourceGroq, Model: g.model,
		Questions: make([]weeklytest.Question, 0, weeklytest.QuestionCount),
	}
	for index, instruction := range batchInstructions {
		generated, err := g.generateBatch(ctx, index+1, instruction, weeklytest.QuestionCount)
		if err != nil {
			return weeklytest.Generated{}, err
		}
		if index == 0 {
			result.Title = generated.Title
			result.Intro = generated.Intro
		}
		for _, question := range generated.Questions {
			result.Questions = append(result.Questions, weeklytest.Question{
				Prompt: question.Prompt, Options: question.Options,
				CorrectIndex: question.CorrectIndex, Explanation: question.Explanation,
				RiskTag: question.RiskTag, Difficulty: question.Difficulty,
			})
		}
	}

	return result, nil
}

func (g *Generator) generateBatch(
	ctx context.Context,
	batchNumber int,
	instruction string,
	expectedQuestions int,
) (generatedPayload, error) {

	payload := chatRequest{
		Model: g.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: instruction},
		},
		Temperature: 0.2,
		// Один запрос на весь экзамен укладывается в бесплатный лимит Groq
		// 8000 TPM. Две части резервировали лимит дважды и давали HTTP 429.
		MaxCompletionTokens: 6200,
		ReasoningEffort:     "low",
		ResponseFormat:      weeklyTestResponseFormat(batchNumber),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return generatedPayload{}, fmt.Errorf("сериализовать запрос Groq: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body),
	)
	if err != nil {
		return generatedPayload{}, fmt.Errorf("создать запрос Groq: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+g.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := g.client.Do(request)
	if err != nil {
		return generatedPayload{}, fmt.Errorf("выполнить запрос Groq: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// Читается только стандартное поле error.message. Заголовок Authorization
		// и отправленный prompt в ошибку никогда не добавляются.
		var apiError struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&apiError)
		detail := strings.Join(strings.Fields(apiError.Error.Message), " ")
		if len(detail) > 500 {
			detail = detail[:500] + "…"
		}
		if detail == "" {
			detail = "без описания"
		}
		return generatedPayload{}, fmt.Errorf(
			"Groq ответил HTTP %d (request-id: %s): %s",
			response.StatusCode,
			response.Header.Get("x-request-id"),
			detail,
		)
	}

	limited := io.LimitReader(response.Body, 1024*1024)
	var decoded chatResponse
	if err := json.NewDecoder(limited).Decode(&decoded); err != nil {
		return generatedPayload{}, fmt.Errorf("разобрать ответ Groq: %w", err)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return generatedPayload{}, fmt.Errorf("Groq вернул пустой ответ")
	}

	var generated generatedPayload
	if err := json.Unmarshal([]byte(decoded.Choices[0].Message.Content), &generated); err != nil {
		return generatedPayload{}, fmt.Errorf("разобрать JSON теста Groq: %w", err)
	}
	if len(generated.Questions) < expectedQuestions {
		return generatedPayload{}, fmt.Errorf(
			"Groq вернул %d вопросов в части %d, ожидалось не меньше %d",
			len(generated.Questions), batchNumber, expectedQuestions,
		)
	}
	// Некоторые модели изредка добавляют лишний вопрос даже при явном требовании
	// точного количества. Лишнее безопасно отбрасывается до доменной валидации.
	generated.Questions = generated.Questions[:expectedQuestions]

	return generated, nil
}

const systemPrompt = `Ты — опытный методист по цифровой безопасности и редактор экзамена для русскоязычного антискам-тренажёра. Твоя цель — проверить практический навык: распознать мошенничество и выбрать действие с минимальным риском.

Требования к качеству:
1. У каждого вопроса ровно 4 правдоподобных варианта и только один однозначно самый безопасный ответ.
2. Правильный вариант не должен содержать даже одного рискованного шага. Он обычно прекращает контакт, предлагает независимую проверку через вручную найденный официальный канал, блокировку/жалобу или обращение в настоящий банк/сервис.
3. Нельзя считать безопасным открытие неожиданного файла после проверки подписи или антивирусом: подпись, антивирус и знакомое расширение не гарантируют безопасность. Не предлагай открывать, скачивать или устанавливать подозрительные файлы.
4. Никогда не делай правильным ответом переход по ссылке или звонок по номеру из подозрительного сообщения, предоплату вне платформы, передачу пароля, кода SMS/2FA, seed-фразы, данных карты, демонстрацию экрана или удалённый доступ.
5. HTTPS, замок, логотип, скриншот, знание имени, голос/видео знакомого и фото документа сами по себе ничего не доказывают. Для проверки личности используй независимый канал и заранее известный контакт.
6. После уже совершённой ошибки безопасный порядок: прекратить контакт, связаться с банком/сервисом официальным способом, заблокировать платёжные средства или сессии, сменить пароли с чистого устройства и сохранить доказательства.
7. Дистракторы должны быть реалистичными, но не равноценными правильному ответу. Не используй шуточные, абсурдные или очевидно преступные варианты.
8. Не делай правильный вариант заметно длиннее остальных. Во всём экзамене используй каждый correctIndex 0, 1, 2 и 3 по пять раз, перемешав порядок.
9. Сюжеты должны отличаться персонажами, каналами, суммами и деталями. Не повторяй одну и ту же проверку под разными формулировками.
10. Соблюдай заданное распределение сложности: easy проверяет базовый красный флаг; medium сочетает 2 признака; hard содержит правдоподобную имитацию и требует правильного порядка нескольких действий.
11. Prompt — конкретная ситуация и один ясный вопрос без недостающих условий. Options — короткие действия в одинаковом стиле.
12. Explanation в 2–3 коротких предложениях: назови конкретный красный флаг, объясни риск выбранного мошенниками действия и дай безопасный порядок действий. Explanation обязано подтверждать correctIndex и не содержать ложных гарантий.
13. riskTag — короткий стабильный код на английском snake_case, описывающий схему.
14. Не используй реальные ссылки, телефоны, реквизиты и персональные данные. Не проси пользователя выполнять действие в реальности.

Перед отправкой JSON молча проведи аудит каждого вопроса:
- безопасен ли correctIndex при всех условиях сюжета;
- совпадает ли Explanation с правильным вариантом;
- нет ли второго столь же безопасного ответа;
- нет ли утверждения, что одиночная техническая проверка гарантирует безопасность.
Если хотя бы один пункт не выполнен — перепиши вопрос до отправки.

Все пользовательские тексты пиши по-русски. Верни только JSON, соответствующий переданной строгой схеме.`

var batchInstructions = []string{
	`Создай полный экзамен: ровно 20 разных вопросов. Покрой: 2 фишинговые ссылки/домены/QR; 2 выдачи себя за банк или родственника и давление срочностью; 1 коды SMS/2FA; 3 мошенничества на маркетплейсах (предоплата, доставка, поддельный чек); 1 поддельная поддержка; 1 удалённый доступ; 1 восстановление или захват аккаунта; 1 вредоносное приложение/вложение; 3 инвестиции/криптовалюта/дроппер/романтическая схема; 2 кражи документов/персональных данных/SIM-swap; 2 современные угрозы (дипфейк, OAuth-фишинг, поддельная реклама или ИИ); 1 безопасная реакция после уже совершённой ошибки. Сложность: 5 easy, 10 medium, 5 hard. Формулируй компактно, чтобы весь экзамен поместился в один ответ.`,
}

func weeklyTestResponseFormat(batchNumber int) map[string]any {
	questionSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"prompt", "options", "correctIndex", "explanation", "riskTag", "difficulty",
		},
		"properties": map[string]any{
			"prompt": map[string]any{"type": "string"},
			"options": map[string]any{
				"type": "array", "minItems": 4, "maxItems": 4,
				"items": map[string]any{"type": "string"},
			},
			"correctIndex": map[string]any{"type": "integer", "minimum": 0, "maximum": 3},
			"explanation":  map[string]any{"type": "string"},
			"riskTag":      map[string]any{"type": "string"},
			"difficulty": map[string]any{
				"type": "string", "enum": []string{"easy", "medium", "hard"},
			},
		},
	}

	return map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   fmt.Sprintf("weekly_antiscam_test_part_%d", batchNumber),
			"strict": true,
			"schema": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"title", "intro", "questions"},
				"properties": map[string]any{
					"title": map[string]any{"type": "string"},
					"intro": map[string]any{"type": "string"},
					"questions": map[string]any{
						"type": "array", "minItems": weeklytest.QuestionCount, "items": questionSchema,
					},
				},
			},
		},
	}
}
