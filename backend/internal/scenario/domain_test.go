package scenario_test

import (
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

func TestValueTypesAcceptOnlySupportedValues(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
		got   bool
	}{
		{name: "роль buyer", valid: true, got: scenario.RoleBuyer.Valid()},
		{name: "роль seller", valid: true, got: scenario.RoleSeller.Valid()},
		{name: "роль admin", valid: false, got: scenario.Role("admin").Valid()},
		{name: "пустая роль", valid: false, got: scenario.Role("").Valid()},

		{name: "сложность easy", valid: true, got: scenario.DifficultyEasy.Valid()},
		{name: "сложность medium", valid: true, got: scenario.DifficultyMedium.Valid()},
		{name: "сложность hard", valid: true, got: scenario.DifficultyHard.Valid()},
		{name: "сложность impossible", valid: false, got: scenario.Difficulty("impossible").Valid()},

		{name: "тип узла message", valid: true, got: scenario.NodeTypeMessage.Valid()},
		{name: "тип узла decision", valid: true, got: scenario.NodeTypeDecision.Valid()},
		{name: "тип узла terminal", valid: true, got: scenario.NodeTypeTerminal.Valid()},
		{name: "тип узла video", valid: false, got: scenario.NodeType("video").Valid()},

		{name: "финал safe", valid: true, got: scenario.OutcomeSafe.Valid()},
		{name: "финал unsafe", valid: true, got: scenario.OutcomeUnsafe.Valid()},
		{name: "финал partial", valid: false, got: scenario.OutcomeType("partial").Valid()},

		{name: "опасность safe", valid: true, got: scenario.SeveritySafe.Valid()},
		{name: "опасность warning", valid: true, got: scenario.SeverityWarning.Valid()},
		{name: "опасность dangerous", valid: true, got: scenario.SeverityDangerous.Valid()},
		{name: "опасность fatal", valid: false, got: scenario.Severity("fatal").Valid()},

		{name: "критичность low", valid: true, got: scenario.CriticalityLow.Valid()},
		{name: "критичность medium", valid: true, got: scenario.CriticalityMedium.Valid()},
		{name: "критичность high", valid: true, got: scenario.CriticalityHigh.Valid()},
		{name: "критичность extreme", valid: false, got: scenario.Criticality("extreme").Valid()},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.got != testCase.valid {
				t.Errorf("Valid() = %v, ожидалось %v", testCase.got, testCase.valid)
			}
		})
	}
}

func TestScenarioNodeLookup(t *testing.T) {
	built, err := scenario.New(validBuyerDraft())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	node, found := built.Node("channel-decision")
	if !found {
		t.Fatal("узел решения должен находиться по идентификатору")
	}

	if node.Type != scenario.NodeTypeDecision {
		t.Errorf("тип узла = %q, ожидался decision", node.Type)
	}

	if _, found := built.Node("unknown-node"); found {
		t.Error("несуществующий узел не должен находиться")
	}
}

func TestNodeChoiceLookup(t *testing.T) {
	built, err := scenario.New(validBuyerDraft())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	node, _ := built.Node("channel-decision")

	choice, found := node.Choice("move-to-messenger")
	if !found {
		t.Fatal("вариант выбора должен находиться по идентификатору")
	}

	if choice.NextNodeID != "pressure-message" {
		t.Errorf("следующий узел = %q, ожидался pressure-message", choice.NextNodeID)
	}

	if _, found := node.Choice("unknown-choice"); found {
		t.Error("несуществующий вариант выбора не должен находиться")
	}
}
