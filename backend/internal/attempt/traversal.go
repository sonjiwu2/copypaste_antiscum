package attempt

import (
	"fmt"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// reveal — результат раскрытия части сценария: подряд идущие сообщения
// и узел, на котором прохождение останавливается.
type reveal struct {
	NodeIDs    []scenario.NodeID
	StopNodeID scenario.NodeID
	Completed  bool
	Outcome    scenario.OutcomeType
}

// revealFrom раскрывает узлы, начиная с указанного, и останавливается
// на первом решении или финале.
//
// Подряд идущие сообщения выдаются одной пачкой: иначе фронтенду пришлось бы
// запрашивать каждую реплику диалога отдельным запросом.
//
// Цикл конечен без искусственного ограничителя, потому что валидатор
// сценариев отвергает графы с циклами ещё при загрузке.
func revealFrom(source scenario.Scenario, from scenario.NodeID) (reveal, error) {
	revealed := reveal{NodeIDs: make([]scenario.NodeID, 0, 4)}
	current := from

	for {
		node, found := source.Node(current)
		if !found {
			return reveal{}, fmt.Errorf("%w: узел %q сценария %q", ErrBrokenScenario, current, source.ID)
		}

		revealed.NodeIDs = append(revealed.NodeIDs, node.ID)

		switch node.Type {
		case scenario.NodeTypeMessage:
			current = node.NextNodeID
		case scenario.NodeTypeDecision:
			revealed.StopNodeID = node.ID

			return revealed, nil
		case scenario.NodeTypeTerminal:
			revealed.StopNodeID = node.ID
			revealed.Completed = true

			if node.TerminalOutcome != nil {
				revealed.Outcome = node.TerminalOutcome.Type
			}

			return revealed, nil
		default:
			return reveal{}, fmt.Errorf("%w: тип узла %q не поддерживается", ErrBrokenScenario, node.Type)
		}
	}
}
