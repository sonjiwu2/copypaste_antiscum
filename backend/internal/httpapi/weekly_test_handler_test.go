package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestWeeklyTestHidesAnswersUntilSubmission(t *testing.T) {
	browser := newTestRouter(t)

	current := httptest.NewRecorder()
	browser.ServeHTTP(current, httptest.NewRequest(
		http.MethodPost, "/api/v1/weekly-tests/current", nil,
	))
	if current.Code != http.StatusOK {
		t.Fatalf("статус создания = %d, ожидался 200: %s", current.Code, current.Body.String())
	}
	if strings.Contains(current.Body.String(), "correctIndex") ||
		strings.Contains(current.Body.String(), "explanation") {
		t.Fatal("непроверенный тест раскрыл правильный ответ")
	}

	var test weeklyTestResponse
	if err := json.Unmarshal(current.Body.Bytes(), &test); err != nil {
		t.Fatalf("разобрать тест: %v", err)
	}
	if test.Source != "fallback" || len(test.Questions) != 20 || test.Completed {
		t.Fatalf("неожиданный тест: %+v", test)
	}

	correctIndexes := []int{2, 1, 3, 1, 2, 2, 1, 1, 2, 2, 1, 2, 1, 2, 2, 1, 2, 1, 1, 2}
	request := submitWeeklyTestRequest{Answers: make([]weeklyTestAnswer, 0, len(correctIndexes))}
	for index, correctIndex := range correctIndexes {
		request.Answers = append(request.Answers, weeklyTestAnswer{
			QuestionID: "q" + strconv.Itoa(index+1), OptionIndex: correctIndex,
		})
	}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("собрать ответы: %v", err)
	}
	submitted := httptest.NewRecorder()
	browser.ServeHTTP(submitted, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/weekly-tests/"+test.TestID+"/submit",
		bytes.NewReader(body),
	))
	if submitted.Code != http.StatusOK {
		t.Fatalf("статус отправки = %d, ожидался 200: %s", submitted.Code, submitted.Body.String())
	}

	var result weeklyTestResponse
	if err := json.Unmarshal(submitted.Body.Bytes(), &result); err != nil {
		t.Fatalf("разобрать результат: %v", err)
	}
	if !result.Completed || result.Result == nil {
		t.Fatal("результат не отмечен завершённым")
	}
	if result.Result.Score != 100 || !result.Result.Passed || result.Result.EarnedXP != 100 || len(result.Result.Review) != 20 {
		t.Fatalf("неверный результат: %+v", result.Result)
	}

	// Повтор не начисляет и не создаёт второй результат.
	repeated := httptest.NewRecorder()
	browser.ServeHTTP(repeated, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/weekly-tests/"+test.TestID+"/submit",
		bytes.NewReader(body),
	))
	if repeated.Code != http.StatusOK || repeated.Body.String() != submitted.Body.String() {
		t.Fatalf("повторная отправка не идемпотентна: %d %s", repeated.Code, repeated.Body.String())
	}
}

func TestWeeklyTestAcceptsPartialAnswersAsFailedExam(t *testing.T) {
	browser := newTestRouter(t)
	current := httptest.NewRecorder()
	browser.ServeHTTP(current, httptest.NewRequest(
		http.MethodPost, "/api/v1/weekly-tests/current", nil,
	))

	var test weeklyTestResponse
	if err := json.Unmarshal(current.Body.Bytes(), &test); err != nil {
		t.Fatalf("разобрать тест: %v", err)
	}

	recorder := httptest.NewRecorder()
	browser.ServeHTTP(recorder, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/weekly-tests/"+test.TestID+"/submit",
		strings.NewReader(`{"answers":[{"questionId":"q1","optionIndex":2}]}`),
	))
	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200: %s", recorder.Code, recorder.Body.String())
	}

	var result weeklyTestResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("разобрать результат: %v", err)
	}
	if result.Result == nil || result.Result.Passed || result.Result.EarnedXP != 0 || result.Result.LivesLeft != 0 {
		t.Fatalf("частичный экзамен должен быть провален: %+v", result.Result)
	}
}
