package scenario_test

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// stubRepository — управляемое хранилище для проверки сервиса без адаптеров.
type stubRepository struct {
	scenarios []scenario.Scenario
	failure   error
}

func (s stubRepository) List(_ context.Context, filter scenario.Filter) ([]scenario.Scenario, error) {
	if s.failure != nil {
		return nil, s.failure
	}

	found := make([]scenario.Scenario, 0, len(s.scenarios))

	for _, stored := range s.scenarios {
		if filter.OnlyActive && !stored.IsActive {
			continue
		}

		if filter.Role != "" && stored.Role != filter.Role {
			continue
		}

		found = append(found, stored)
	}

	return found, nil
}

func (s stubRepository) Get(_ context.Context, id scenario.ID) (scenario.Scenario, error) {
	if s.failure != nil {
		return scenario.Scenario{}, s.failure
	}

	for _, stored := range s.scenarios {
		if stored.ID == id {
			return stored, nil
		}
	}

	return scenario.Scenario{}, scenario.ErrNotFound
}

func catalogRepository(t *testing.T) stubRepository {
	t.Helper()

	buyer, err := scenario.New(validBuyerDraft())
	if err != nil {
		t.Fatalf("сценарий покупателя невалиден: %v", err)
	}

	seller, err := scenario.New(validSellerDraft())
	if err != nil {
		t.Fatalf("сценарий продавца невалиден: %v", err)
	}

	hiddenDraft := validBuyerDraft()
	hiddenDraft.ID = "hidden-scenario"
	hiddenDraft.IsActive = false

	hidden, err := scenario.New(hiddenDraft)
	if err != nil {
		t.Fatalf("скрытый сценарий невалиден: %v", err)
	}

	// Порядок намеренно перемешан: сервис обязан упорядочить каталог сам.
	return stubRepository{scenarios: []scenario.Scenario{seller, hidden, buyer}}
}

func TestServiceList(t *testing.T) {
	testCases := []struct {
		name    string
		role    scenario.Role
		wantIDs []scenario.ID
		wantErr error
	}{
		{
			name:    "все активные сценарии",
			role:    "",
			wantIDs: []scenario.ID{"buyer-fake-delivery", "seller-payment-already-sent"},
		},
		{
			name:    "только покупатель",
			role:    scenario.RoleBuyer,
			wantIDs: []scenario.ID{"buyer-fake-delivery"},
		},
		{
			name:    "только продавец",
			role:    scenario.RoleSeller,
			wantIDs: []scenario.ID{"seller-payment-already-sent"},
		},
		{
			name:    "неподдерживаемая роль",
			role:    "courier",
			wantErr: scenario.ErrUnsupportedRole,
		},
	}

	service := scenario.NewService(catalogRepository(t))

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			catalog, err := service.List(context.Background(), testCase.role)

			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("ошибка = %v, ожидалась %v", err, testCase.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			gotIDs := make([]scenario.ID, len(catalog))
			for i, item := range catalog {
				gotIDs[i] = item.ID
			}

			if !reflect.DeepEqual(gotIDs, testCase.wantIDs) {
				t.Errorf("каталог = %v, ожидался %v", gotIDs, testCase.wantIDs)
			}
		})
	}
}

// Клиент видит один и тот же порядок при одинаковом запросе.
func TestServiceListIsOrderedDeterministically(t *testing.T) {
	const attempts = 15

	service := scenario.NewService(catalogRepository(t))

	var first []scenario.ID

	for range attempts {
		catalog, err := service.List(context.Background(), "")
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		current := make([]scenario.ID, len(catalog))
		for i, item := range catalog {
			current[i] = item.ID
		}

		if first == nil {
			first = current

			continue
		}

		if !reflect.DeepEqual(current, first) {
			t.Fatalf("порядок нестабилен: %v и %v", first, current)
		}
	}

	if !sort.SliceIsSorted(first, func(i, j int) bool { return first[i] < first[j] }) {
		t.Errorf("каталог не отсортирован: %v", first)
	}
}

func TestServiceListPropagatesRepositoryFailure(t *testing.T) {
	storageFailure := errors.New("хранилище недоступно")
	service := scenario.NewService(stubRepository{failure: storageFailure})

	if _, err := service.List(context.Background(), ""); !errors.Is(err, storageFailure) {
		t.Fatalf("ошибка = %v, ожидалась обёртка над %v", err, storageFailure)
	}
}

func TestServiceGet(t *testing.T) {
	testCases := []struct {
		name      string
		id        scenario.ID
		wantErr   error
		wantTitle string
	}{
		{
			name:      "существующий активный сценарий",
			id:        "buyer-fake-delivery",
			wantTitle: "Ссылка на доставку",
		},
		{
			name:    "неизвестный сценарий",
			id:      "unknown-scenario",
			wantErr: scenario.ErrNotFound,
		},
		{
			name:    "отключённый сценарий выглядит отсутствующим",
			id:      "hidden-scenario",
			wantErr: scenario.ErrNotFound,
		},
	}

	service := scenario.NewService(catalogRepository(t))

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			metadata, err := service.Get(context.Background(), testCase.id)

			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("ошибка = %v, ожидалась %v", err, testCase.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if metadata.Title != testCase.wantTitle {
				t.Errorf("заголовок = %q, ожидался %q", metadata.Title, testCase.wantTitle)
			}
		})
	}
}

// Проекция каталога не должна обзавестись полями графа: узлы, выборы и
// последствия остаются на сервере до начала попытки.
func TestMetadataExposesOnlyCatalogFields(t *testing.T) {
	allowed := map[string]struct{}{
		"ID": {}, "Version": {}, "Slug": {}, "Role": {},
		"Title": {}, "Description": {}, "Difficulty": {}, "EstimatedMinutes": {},
	}

	metadataType := reflect.TypeOf(scenario.Metadata{})

	for i := range metadataType.NumField() {
		name := metadataType.Field(i).Name

		if _, ok := allowed[name]; !ok {
			t.Errorf("в Metadata появилось поле %q — проверьте, не утекает ли граф сценария", name)
		}
	}

	if metadataType.NumField() != len(allowed) {
		t.Errorf("полей в Metadata = %d, ожидалось %d", metadataType.NumField(), len(allowed))
	}
}
