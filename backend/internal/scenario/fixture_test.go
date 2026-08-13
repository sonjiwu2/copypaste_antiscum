package scenario_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/scenarios"
)

// minimalFixture — самый маленький корректный файл сценария.
const minimalFixture = `{
  "id": "demo",
  "version": 1,
  "slug": "demo",
  "role": "buyer",
  "title": "Демонстрация",
  "description": "Проверочный сценарий.",
  "difficulty": "easy",
  "estimatedMinutes": 1,
  "startNodeId": "start",
  "isActive": true,
  "nodes": [
    { "id": "start", "type": "message", "sender": "seller", "text": "Привет.", "nextNodeId": "pick" },
    {
      "id": "pick",
      "type": "decision",
      "decisionPrompt": "Что делать?",
      "choices": [
        {
          "id": "safe", "label": "Безопасно", "playerReply": "Так делать не буду.", "nextNodeId": "good",
          "safetyScore": 0, "criticality": "low",
          "consequence": { "severity": "safe", "title": "Верно", "explanation": "Так безопаснее.", "realWorldRule": "Правило." }
        },
        {
          "id": "risky", "label": "Опасно", "playerReply": "Хорошо, согласен.", "nextNodeId": "bad",
          "safetyScore": -10, "criticality": "high",
          "consequence": { "severity": "dangerous", "title": "Ошибка", "explanation": "Так делать нельзя.", "realWorldRule": "Правило." }
        }
      ]
    },
    { "id": "good", "type": "terminal", "outcome": { "type": "safe", "title": "Хорошо", "explanation": "Всё в порядке." } },
    { "id": "bad", "type": "terminal", "outcome": { "type": "unsafe", "title": "Плохо", "explanation": "Деньги потеряны." } }
  ]
}`

func TestLoadFSReadsEmbeddedScenarios(t *testing.T) {
	loaded, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("встроенные сценарии должны загружаться: %v", err)
	}

	if len(loaded) < 2 {
		t.Fatalf("загружено %d сценариев, ожидалось минимум 2", len(loaded))
	}

	roles := make(map[scenario.Role]int, len(loaded))

	for _, found := range loaded {
		roles[found.Role]++

		if found.Version < 1 {
			t.Errorf("сценарий %q: версия %d, ожидалась положительная", found.ID, found.Version)
		}

		if !found.IsActive {
			t.Errorf("сценарий %q должен быть активным", found.ID)
		}

		if _, exists := found.Node(found.StartNodeID); !exists {
			t.Errorf("сценарий %q: стартовый узел не найден", found.ID)
		}
	}

	if roles[scenario.RoleBuyer] == 0 {
		t.Error("нужен хотя бы один сценарий покупателя")
	}

	if roles[scenario.RoleSeller] == 0 {
		t.Error("нужен хотя бы один сценарий продавца")
	}
}

// Каждая фикстура должна доводить и до безопасного, и до небезопасного финала,
// иначе тренажёр не показывает последствия ошибки.
func TestEmbeddedScenariosHaveBothEndings(t *testing.T) {
	loaded, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	for _, found := range loaded {
		outcomes := map[scenario.OutcomeType]int{}
		decisions := 0

		for _, nodeID := range collectNodeIDs(t, found) {
			node, _ := found.Node(nodeID)

			switch node.Type {
			case scenario.NodeTypeDecision:
				decisions++
			case scenario.NodeTypeTerminal:
				outcomes[node.TerminalOutcome.Type]++
			case scenario.NodeTypeMessage:
			}
		}

		if outcomes[scenario.OutcomeSafe] == 0 || outcomes[scenario.OutcomeUnsafe] == 0 {
			t.Errorf("сценарий %q: нужны безопасный и небезопасный финалы, получено %v", found.ID, outcomes)
		}

		if decisions < 3 {
			t.Errorf("сценарий %q: решений %d, ожидалось минимум 3", found.ID, decisions)
		}
	}
}

// В сценариях покупателя разговор начинается с его вопроса, а не с ответа
// продавца на невидимую реплику. В сценариях продавца первый обычный ответ на
// приветствие покупателя также не должен оставаться за кадром.
func TestEmbeddedScenariosKeepOpeningPlayerLines(t *testing.T) {
	loaded, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	for _, found := range loaded {
		start, ok := found.Node(found.StartNodeID)
		if !ok {
			t.Errorf("сценарий %q: стартовый узел не найден", found.ID)
			continue
		}

		if found.Role == scenario.RoleBuyer {
			if start.Type != scenario.NodeTypeMessage || start.Sender != string(found.Role) {
				t.Errorf("сценарий %q: ожидался стартовый вопрос покупателя, получено %s:%s",
					found.ID, start.Type, start.Sender)
			}
			continue
		}

		next, ok := found.Node(start.NextNodeID)
		if ok && next.Type == scenario.NodeTypeMessage && next.Sender != string(found.Role) {
			t.Errorf("сценарий %q: после приветствия покупателя пропущен ответ продавца", found.ID)
		}
	}
}

func TestMacBookOpeningContainsBothSidesOfConversation(t *testing.T) {
	loaded, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	var macbook scenario.Scenario
	for _, found := range loaded {
		if found.ID == "buyer-macbook-corporate-lock" {
			macbook = found
			break
		}
	}
	if macbook.ID == "" {
		t.Fatal("сценарий MacBook не найден")
	}

	wantSenders := []string{"buyer", "seller", "buyer", "seller"}
	current := macbook.StartNodeID
	for index, wantSender := range wantSenders {
		node, ok := macbook.Node(current)
		if !ok {
			t.Fatalf("реплика %d: узел %q не найден", index+1, current)
		}
		if node.Type != scenario.NodeTypeMessage || node.Sender != wantSender || strings.TrimSpace(node.Text) == "" {
			t.Errorf("реплика %d: получено %s:%s %q, ожидался непустой message:%s",
				index+1, node.Type, node.Sender, node.Text, wantSender)
		}
		current = node.NextNodeID
	}
}

func TestEmbeddedScenariosContainScriptedBridgeLines(t *testing.T) {
	loaded, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	byID := make(map[scenario.ID]scenario.Scenario, len(loaded))
	for _, found := range loaded {
		byID[found.ID] = found
	}

	testCases := []struct {
		scenarioID scenario.ID
		nodeID     scenario.NodeID
		sender     string
		textPart   string
	}{
		{"buyer-gpu-hidden-repair", "buyer-video-meeting-question", "buyer", "проверить карту"},
		{"buyer-gpu-hidden-repair", "buyer-history-meeting-question", "buyer", "Когда можно подъехать"},
		{"buyer-iphone-deposit", "buyer-documents-objection", "buyer", "На чужую карту"},
		{"buyer-ps5-delivery", "buyer-support-city", "buyer", "Нижний Новгород"},
		{"buyer-ps5-delivery", "buyer-pvz-choice", "buyer", "улице Белинского"},
		{"seller-laptop-courier", "seller-requires-confirmed-payment", "seller", "банковском приложении"},
		{"seller-laptop-courier", "seller-repeats-payment-rule", "seller", "товар я не передам"},
	}

	for _, testCase := range testCases {
		built, ok := byID[testCase.scenarioID]
		if !ok {
			t.Errorf("сценарий %q не найден", testCase.scenarioID)
			continue
		}

		node, ok := built.Node(testCase.nodeID)
		if !ok {
			t.Errorf("сценарий %q: связующая реплика %q не найдена", testCase.scenarioID, testCase.nodeID)
			continue
		}
		if node.Type != scenario.NodeTypeMessage || node.Sender != testCase.sender ||
			!strings.Contains(node.Text, testCase.textPart) {
			t.Errorf("сценарий %q, узел %q: получено %s:%s %q",
				testCase.scenarioID, testCase.nodeID, node.Type, node.Sender, node.Text)
		}
	}
}

func TestLoadFSRejectsBrokenFixtures(t *testing.T) {
	testCases := []struct {
		name        string
		files       fstest.MapFS
		wantMessage string
	}{
		{
			name:        "нет ни одного файла",
			files:       fstest.MapFS{},
			wantMessage: "не найдено ни одного файла сценария",
		},
		{
			name:        "битый JSON",
			files:       fstest.MapFS{"broken.json": &fstest.MapFile{Data: []byte("{ это не json")}},
			wantMessage: "разобрать файл сценария",
		},
		{
			name: "неизвестное поле",
			files: fstest.MapFS{"typo.json": &fstest.MapFile{
				Data: []byte(strings.Replace(minimalFixture, `"slug": "demo"`, `"slugg": "demo"`, 1)),
			}},
			wantMessage: "разобрать файл сценария",
		},
		{
			name: "граф не проходит проверку",
			files: fstest.MapFS{"invalid.json": &fstest.MapFile{
				Data: []byte(strings.Replace(minimalFixture, `"nextNodeId": "good"`, `"nextNodeId": "ghost"`, 1)),
			}},
			wantMessage: "не прошёл проверку",
		},
		{
			name: "неподдерживаемая роль",
			files: fstest.MapFS{"role.json": &fstest.MapFile{
				Data: []byte(strings.Replace(minimalFixture, `"role": "buyer"`, `"role": "courier"`, 1)),
			}},
			wantMessage: "не прошёл проверку",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			loaded, err := scenario.LoadFS(testCase.files)
			if err == nil {
				t.Fatalf("ожидалась ошибка, загружено %d сценариев", len(loaded))
			}

			if !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Errorf("сообщение %q не содержит %q", err.Error(), testCase.wantMessage)
			}
		})
	}
}

func TestLoadFSAcceptsMinimalFixture(t *testing.T) {
	loaded, err := scenario.LoadFS(fstest.MapFS{
		"demo.json": &fstest.MapFile{Data: []byte(minimalFixture)},
	})
	if err != nil {
		t.Fatalf("минимальная фикстура должна загружаться: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("загружено %d сценариев, ожидался 1", len(loaded))
	}

	node, found := loaded[0].Node("pick")
	if !found {
		t.Fatal("узел решения должен быть загружен")
	}

	choice, found := node.Choice("risky")
	if !found {
		t.Fatal("вариант выбора должен быть загружен")
	}

	if choice.SafetyScore != -10 {
		t.Errorf("SafetyScore = %d, ожидалось -10", choice.SafetyScore)
	}

	if choice.Consequence.Severity != scenario.SeverityDangerous {
		t.Errorf("severity = %q, ожидалось dangerous", choice.Consequence.Severity)
	}
}

// collectNodeIDs обходит сценарий от старта, чтобы тест не зависел
// от внутреннего представления набора узлов.
func collectNodeIDs(t *testing.T, built scenario.Scenario) []scenario.NodeID {
	t.Helper()

	visited := map[scenario.NodeID]struct{}{}
	queue := []scenario.NodeID{built.StartNodeID}
	order := make([]scenario.NodeID, 0, built.NodeCount())

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if _, seen := visited[current]; seen {
			continue
		}

		node, found := built.Node(current)
		if !found {
			continue
		}

		visited[current] = struct{}{}
		order = append(order, current)

		if node.NextNodeID != "" {
			queue = append(queue, node.NextNodeID)
		}

		for _, choice := range node.Choices {
			queue = append(queue, choice.NextNodeID)
		}
	}

	return order
}
