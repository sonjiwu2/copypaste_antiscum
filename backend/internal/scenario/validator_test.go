package scenario_test

import (
	"errors"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// validBuyerDraft возвращает корректный сценарий покупателя.
// Каждый тест получает свежую копию и портит ровно одно правило.
func validBuyerDraft() scenario.Draft {
	return scenario.Draft{
		ID:               "buyer-fake-delivery",
		Version:          1,
		Slug:             "buyer-fake-delivery",
		Role:             scenario.RoleBuyer,
		Title:            "Ссылка на доставку",
		Description:      "Учебный сценарий покупателя.",
		Difficulty:       scenario.DifficultyMedium,
		EstimatedMinutes: 4,
		StartNodeID:      "greeting",
		IsActive:         true,
		Nodes: []scenario.Node{
			{
				ID:         "greeting",
				Type:       scenario.NodeTypeMessage,
				Sender:     "seller",
				Text:       "Здравствуйте, товар ещё доступен.",
				NextNodeID: "channel-decision",
			},
			{
				ID:             "channel-decision",
				Type:           scenario.NodeTypeDecision,
				DecisionPrompt: "Что вы сделаете?",
				Choices: []scenario.Choice{
					{
						ID:          "stay-on-platform",
						Label:       "Продолжить в приложении",
						NextNodeID:  "safe-ending",
						SafetyScore: 0,
						Criticality: scenario.CriticalityLow,
						RiskTags:    []scenario.RiskTag{"off_platform"},
						SkillEffects: []scenario.SkillEffect{
							{Skill: "channel_safety", Delta: 1},
						},
						Consequence: scenario.Consequence{
							Severity:      scenario.SeveritySafe,
							Title:         "Переписка сохранена",
							Explanation:   "История общения остаётся внутри площадки.",
							RealWorldRule: "Не уводите общение в сторонний мессенджер.",
						},
					},
					{
						ID:          "move-to-messenger",
						Label:       "Перейти в мессенджер",
						NextNodeID:  "pressure-message",
						SafetyScore: -20,
						Criticality: scenario.CriticalityHigh,
						RiskTags:    []scenario.RiskTag{"off_platform"},
						Consequence: scenario.Consequence{
							Severity:      scenario.SeverityDangerous,
							Title:         "Защита площадки потеряна",
							Explanation:   "Вне площадки переписка не поможет вернуть деньги.",
							RealWorldRule: "Оставайтесь внутри сервиса до конца сделки.",
						},
					},
				},
			},
			{
				ID:         "pressure-message",
				Type:       scenario.NodeTypeMessage,
				Sender:     "seller",
				Text:       "Оплатите по ссылке, иначе бронь снимут.",
				NextNodeID: "unsafe-ending",
			},
			{
				ID:   "safe-ending",
				Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{
					Type:        scenario.OutcomeSafe,
					Title:       "Сделка прошла безопасно",
					Explanation: "Вы остались в защищённом канале.",
				},
			},
			{
				ID:   "unsafe-ending",
				Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{
					Type:        scenario.OutcomeUnsafe,
					Title:       "Деньги потеряны",
					Explanation: "Оплата ушла мошеннику вне площадки.",
				},
			},
		},
	}
}

// validSellerDraft возвращает корректный сценарий продавца.
func validSellerDraft() scenario.Draft {
	return scenario.Draft{
		ID:               "seller-payment-already-sent",
		Version:          1,
		Slug:             "seller-payment-already-sent",
		Role:             scenario.RoleSeller,
		Title:            "Оплата уже отправлена",
		Description:      "Учебный сценарий продавца.",
		Difficulty:       scenario.DifficultyHard,
		EstimatedMinutes: 5,
		StartNodeID:      "buyer-message",
		IsActive:         true,
		Nodes: []scenario.Node{
			{
				ID:         "buyer-message",
				Type:       scenario.NodeTypeMessage,
				Sender:     "buyer",
				Text:       "Я оплатил, пришлите код из СМС для подтверждения.",
				NextNodeID: "code-decision",
			},
			{
				ID:             "code-decision",
				Type:           scenario.NodeTypeDecision,
				DecisionPrompt: "Что вы ответите?",
				Choices: []scenario.Choice{
					{
						ID:          "refuse-code",
						Label:       "Отказаться называть код",
						NextNodeID:  "safe-ending",
						SafetyScore: 0,
						Criticality: scenario.CriticalityLow,
						Consequence: scenario.Consequence{
							Severity:      scenario.SeveritySafe,
							Title:         "Доступ к счёту сохранён",
							Explanation:   "Код из СМС подтверждает операцию, а не оплату.",
							RealWorldRule: "Никому не сообщайте код из СМС.",
						},
					},
					{
						ID:          "send-code",
						Label:       "Отправить код",
						NextNodeID:  "unsafe-ending",
						SafetyScore: -40,
						Criticality: scenario.CriticalityHigh,
						Consequence: scenario.Consequence{
							Severity:      scenario.SeverityDangerous,
							Title:         "Счёт под угрозой",
							Explanation:   "Кодом подтверждают списание, а не поступление денег.",
							RealWorldRule: "Код из СМС — это подпись под операцией.",
						},
					},
				},
			},
			{
				ID:   "safe-ending",
				Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{
					Type:        scenario.OutcomeSafe,
					Title:       "Вы сохранили доступ",
					Explanation: "Код остался только у вас.",
				},
			},
			{
				ID:   "unsafe-ending",
				Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{
					Type:        scenario.OutcomeUnsafe,
					Title:       "Деньги списаны",
					Explanation: "Код позволил подтвердить чужую операцию.",
				},
			},
		},
	}
}

func TestNewAcceptsValidScenarios(t *testing.T) {
	testCases := []struct {
		name  string
		draft scenario.Draft
	}{
		{name: "корректный сценарий покупателя", draft: validBuyerDraft()},
		{name: "корректный сценарий продавца", draft: validSellerDraft()},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			built, err := scenario.New(testCase.draft)
			if err != nil {
				t.Fatalf("сценарий должен быть валидным, получена ошибка: %v", err)
			}

			if built.ID != testCase.draft.ID {
				t.Errorf("ID = %q, ожидался %q", built.ID, testCase.draft.ID)
			}

			if built.NodeCount() != len(testCase.draft.Nodes) {
				t.Errorf("узлов = %d, ожидалось %d", built.NodeCount(), len(testCase.draft.Nodes))
			}

			if _, found := built.Node(built.StartNodeID); !found {
				t.Error("стартовый узел должен быть доступен")
			}
		})
	}
}

func TestNewRejectsInvalidScenarios(t *testing.T) {
	testCases := []struct {
		name     string
		mutate   func(draft *scenario.Draft)
		wantRule scenario.ValidationRule
	}{
		{
			name:     "пустой идентификатор сценария",
			mutate:   func(d *scenario.Draft) { d.ID = "" },
			wantRule: scenario.RuleScenarioIDRequired,
		},
		{
			name:     "неположительная версия",
			mutate:   func(d *scenario.Draft) { d.Version = 0 },
			wantRule: scenario.RuleScenarioVersionInvalid,
		},
		{
			name:     "неподдерживаемая роль",
			mutate:   func(d *scenario.Draft) { d.Role = "moderator" },
			wantRule: scenario.RuleScenarioRoleUnsupported,
		},
		{
			name:     "пустой заголовок",
			mutate:   func(d *scenario.Draft) { d.Title = "" },
			wantRule: scenario.RuleScenarioTitleRequired,
		},
		{
			name:     "неподдерживаемая сложность",
			mutate:   func(d *scenario.Draft) { d.Difficulty = "impossible" },
			wantRule: scenario.RuleScenarioDifficultyUnsupported,
		},
		{
			name:     "стартовый узел не существует",
			mutate:   func(d *scenario.Draft) { d.StartNodeID = "unknown-node" },
			wantRule: scenario.RuleStartNodeMissing,
		},
		{
			name: "повторяющийся идентификатор узла",
			mutate: func(d *scenario.Draft) {
				d.Nodes = append(d.Nodes, scenario.Node{
					ID:         "greeting",
					Type:       scenario.NodeTypeMessage,
					Text:       "Дубликат.",
					NextNodeID: "safe-ending",
				})
			},
			wantRule: scenario.RuleDuplicateNodeID,
		},
		{
			name: "узел без идентификатора",
			mutate: func(d *scenario.Draft) {
				d.Nodes = append(d.Nodes, scenario.Node{
					Type:       scenario.NodeTypeMessage,
					Text:       "Безымянный узел.",
					NextNodeID: "safe-ending",
				})
			},
			wantRule: scenario.RuleNodeIDRequired,
		},
		{
			name:     "неподдерживаемый тип узла",
			mutate:   func(d *scenario.Draft) { d.Nodes[0].Type = "video" },
			wantRule: scenario.RuleNodeTypeUnsupported,
		},
		{
			name:     "сообщение без следующего узла",
			mutate:   func(d *scenario.Draft) { d.Nodes[0].NextNodeID = "" },
			wantRule: scenario.RuleMessageNodeWithoutNext,
		},
		{
			name: "сообщение с вариантами выбора",
			mutate: func(d *scenario.Draft) {
				d.Nodes[0].Choices = []scenario.Choice{{ID: "any", Label: "Вариант", NextNodeID: "safe-ending"}}
			},
			wantRule: scenario.RuleMessageNodeWithChoices,
		},
		{
			name:     "решение с единственным вариантом",
			mutate:   func(d *scenario.Draft) { d.Nodes[1].Choices = d.Nodes[1].Choices[:1] },
			wantRule: scenario.RuleDecisionNodeTooFewChoices,
		},
		{
			name:     "повторяющийся идентификатор выбора",
			mutate:   func(d *scenario.Draft) { d.Nodes[1].Choices[1].ID = d.Nodes[1].Choices[0].ID },
			wantRule: scenario.RuleDuplicateChoiceID,
		},
		{
			name:     "выбор без подписи",
			mutate:   func(d *scenario.Draft) { d.Nodes[1].Choices[0].Label = "" },
			wantRule: scenario.RuleChoiceLabelRequired,
		},
		{
			name:     "выбор ведёт в несуществующий узел",
			mutate:   func(d *scenario.Draft) { d.Nodes[1].Choices[0].NextNodeID = "ghost-node" },
			wantRule: scenario.RuleTransitionTargetMissing,
		},
		{
			name:     "сообщение ведёт в несуществующий узел",
			mutate:   func(d *scenario.Draft) { d.Nodes[0].NextNodeID = "ghost-node" },
			wantRule: scenario.RuleTransitionTargetMissing,
		},
		{
			name:     "терминал ссылается на следующий узел",
			mutate:   func(d *scenario.Draft) { d.Nodes[3].NextNodeID = "unsafe-ending" },
			wantRule: scenario.RuleTerminalNodeWithNext,
		},
		{
			name: "терминал содержит варианты выбора",
			mutate: func(d *scenario.Draft) {
				d.Nodes[3].Choices = []scenario.Choice{{ID: "any", Label: "Вариант", NextNodeID: "unsafe-ending"}}
			},
			wantRule: scenario.RuleTerminalNodeWithChoices,
		},
		{
			name:     "терминал без итога",
			mutate:   func(d *scenario.Draft) { d.Nodes[3].TerminalOutcome = nil },
			wantRule: scenario.RuleTerminalNodeWithoutOutcome,
		},
		{
			name:     "итог с неподдерживаемым типом",
			mutate:   func(d *scenario.Draft) { d.Nodes[3].TerminalOutcome.Type = "unknown" },
			wantRule: scenario.RuleTerminalNodeWithoutOutcome,
		},
		{
			name:     "последствие без объяснения",
			mutate:   func(d *scenario.Draft) { d.Nodes[1].Choices[0].Consequence.Explanation = "" },
			wantRule: scenario.RuleConsequenceIncomplete,
		},
		{
			name:     "последствие с неподдерживаемой опасностью",
			mutate:   func(d *scenario.Draft) { d.Nodes[1].Choices[0].Consequence.Severity = "fatal" },
			wantRule: scenario.RuleConsequenceIncomplete,
		},
		{
			name: "недостижимый узел",
			mutate: func(d *scenario.Draft) {
				d.Nodes = append(d.Nodes, scenario.Node{
					ID:         "orphan-message",
					Type:       scenario.NodeTypeMessage,
					Text:       "Этот узел никто не открывает.",
					NextNodeID: "safe-ending",
				})
			},
			wantRule: scenario.RuleNodeUnreachable,
		},
		{
			name: "цикл в графе",
			mutate: func(d *scenario.Draft) {
				d.Nodes[2].NextNodeID = "channel-decision"
			},
			wantRule: scenario.RuleGraphHasCycle,
		},
		{
			name: "недостижимый финал",
			mutate: func(d *scenario.Draft) {
				d.Nodes[1].Choices[0].NextNodeID = "pressure-message"
				d.Nodes[2].NextNodeID = "ghost-node"
			},
			wantRule: scenario.RuleTerminalUnreachable,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			draft := validBuyerDraft()
			testCase.mutate(&draft)

			built, err := scenario.New(draft)
			if err == nil {
				t.Fatal("ожидалась ошибка валидации, сценарий принят")
			}

			if built.NodeCount() != 0 {
				t.Error("невалидный сценарий не должен возвращаться вызывающему коду")
			}

			var violations scenario.ValidationErrors
			if !errors.As(err, &violations) {
				t.Fatalf("ожидался тип ValidationErrors, получено %T", err)
			}

			if !violations.Has(testCase.wantRule) {
				t.Errorf("правило %q не сработало, найдены: %v", testCase.wantRule, violations.Rules())
			}
		})
	}
}

func TestValidationErrorCarriesLocation(t *testing.T) {
	draft := validBuyerDraft()
	draft.Nodes[1].Choices[0].Label = ""

	_, err := scenario.New(draft)

	var violations scenario.ValidationErrors
	if !errors.As(err, &violations) {
		t.Fatalf("ожидался тип ValidationErrors, получено %T", err)
	}

	for _, violation := range violations {
		if violation.Rule != scenario.RuleChoiceLabelRequired {
			continue
		}

		if violation.ScenarioID != draft.ID {
			t.Errorf("ScenarioID = %q, ожидался %q", violation.ScenarioID, draft.ID)
		}

		if violation.NodeID != "channel-decision" {
			t.Errorf("NodeID = %q, ожидался \"channel-decision\"", violation.NodeID)
		}

		if violation.ChoiceID != "stay-on-platform" {
			t.Errorf("ChoiceID = %q, ожидался \"stay-on-platform\"", violation.ChoiceID)
		}

		if violation.Error() == "" {
			t.Error("описание нарушения не должно быть пустым")
		}

		return
	}

	t.Fatal("нарушение с правилом choice_label_required не найдено")
}

func TestCycleDetectionIsDeterministic(t *testing.T) {
	const attempts = 20

	draft := validBuyerDraft()
	draft.Nodes[2].NextNodeID = "channel-decision"

	var firstDetail string

	for range attempts {
		_, err := scenario.New(draft)

		var violations scenario.ValidationErrors
		if !errors.As(err, &violations) {
			t.Fatalf("ожидался тип ValidationErrors, получено %T", err)
		}

		detail := ""

		for _, violation := range violations {
			if violation.Rule == scenario.RuleGraphHasCycle {
				detail = string(violation.NodeID)
			}
		}

		if detail == "" {
			t.Fatal("цикл не обнаружен")
		}

		if firstDetail == "" {
			firstDetail = detail

			continue
		}

		if detail != firstDetail {
			t.Fatalf("узел цикла нестабилен: %q и %q", firstDetail, detail)
		}
	}
}
