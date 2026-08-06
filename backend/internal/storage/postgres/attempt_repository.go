package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
)

// ErrHistoryRewrite — сохранённая история решений не совпадает с историей
// переданного агрегата.
//
// Решения неизменяемы: расхождение означает дефект приложения, а не
// конкурентный запрос, поэтому такая ошибка не превращается в конфликт.
var ErrHistoryRewrite = errors.New("история решений попытки неизменяема")

// AttemptRepository хранит попытки в PostgreSQL.
//
// Адаптер повторяет поведение in-memory реализации: те же доменные ошибки,
// та же семантика версий и та же независимость возвращаемых агрегатов.
// Роль оптимистичной блокировки играет условие version в UPDATE.
type AttemptRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewAttemptRepository создаёт хранилище попыток поверх пула подключений.
func NewAttemptRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *AttemptRepository {
	return &AttemptRepository{pool: pool, queryTimeout: queryTimeout}
}

const insertAttemptQuery = `
INSERT INTO attempts (
    id, profile_id, scenario_id, scenario_version, current_node_id, status, score,
    outcome, started_at, updated_at, completed_at, revealed_node_ids,
    applied_skill_effects, version
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

const selectAttemptQuery = `
SELECT id, profile_id, scenario_id, scenario_version, current_node_id, status, score,
       outcome, started_at, updated_at, completed_at, revealed_node_ids,
       applied_skill_effects, version
  FROM attempts
 WHERE id = $1`

const selectDecisionsQuery = `
SELECT ordinal, node_id, choice_id, choice_label, idempotency_key, consequence,
       criticality, risk_tags, skill_effects, score_delta, created_at,
       revealed_node_ids, resulting_node_id, completed, outcome, score_after
  FROM attempt_decisions
 WHERE attempt_id = $1
 ORDER BY ordinal ASC`

// updateAttemptQuery — ядро оптимистичной блокировки. Условие по версии и
// её увеличение выполняются одним запросом, поэтому два параллельных
// перехода не могут примениться оба.
const updateAttemptQuery = `
UPDATE attempts
   SET current_node_id = $2,
       status = $3,
       score = $4,
       outcome = $5,
       updated_at = $6,
       completed_at = $7,
       revealed_node_ids = $8,
       applied_skill_effects = $9,
       version = version + 1
 WHERE id = $1
   AND version = $10
RETURNING version`

const selectAttemptExistsQuery = `SELECT 1 FROM attempts WHERE id = $1`

const countDecisionsQuery = `SELECT count(*) FROM attempt_decisions WHERE attempt_id = $1`

const selectLastDecisionKeyQuery = `
SELECT idempotency_key
  FROM attempt_decisions
 WHERE attempt_id = $1
 ORDER BY ordinal DESC
 LIMIT 1`

const insertDecisionQuery = `
INSERT INTO attempt_decisions (
    attempt_id, ordinal, node_id, choice_id, choice_label, idempotency_key,
    consequence, criticality, risk_tags, skill_effects, score_delta, created_at,
    revealed_node_ids, resulting_node_id, completed, outcome, score_after
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

// Create сохраняет новую попытку вместе с её решениями.
//
// Вставка идёт одной транзакцией: попытка без своих решений — уже испорченная
// история, даже если решений при старте обычно нет.
func (r *AttemptRepository) Create(ctx context.Context, created attempt.Attempt) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validateForPersistence(created); err != nil {
		return err
	}

	row, err := attemptRowOf(created)
	if err != nil {
		return err
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	return r.inTransaction(queryCtx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		_, err := tx.Exec(queryCtx, insertAttemptQuery,
			row.ID, row.ProfileID, row.ScenarioID, row.ScenarioVersion, row.CurrentNodeID,
			row.Status, row.Score, row.Outcome, row.StartedAt, row.UpdatedAt, row.CompletedAt,
			row.RevealedNodeIDs, row.AppliedSkillEffects, row.Version)
		if err != nil {
			return fmt.Errorf("сохранить попытку: %w", MapError(err))
		}

		return r.insertDecisions(queryCtx, tx, created.ID, created.Decisions, 0)
	})
}

// Get возвращает попытку целиком.
//
// Чтение идёт в одной транзакции с повторяемым чтением: строка попытки и её
// решения обязаны прийти из одного состояния базы, иначе параллельный переход
// мог бы отдать клиенту историю новее, чем сама попытка.
func (r *AttemptRepository) Get(ctx context.Context, id attempt.ID) (attempt.Attempt, error) {
	if err := ctx.Err(); err != nil {
		return attempt.Attempt{}, err
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var found attempt.Attempt

	err := r.inTransaction(queryCtx, pgx.RepeatableRead, func(tx pgx.Tx) error {
		loaded, err := r.loadAttempt(queryCtx, tx, id)
		if err != nil {
			return err
		}

		found = loaded

		return nil
	})
	if err != nil {
		return attempt.Attempt{}, err
	}

	return found, nil
}

// Update сохраняет переход попытки, если её версия не устарела.
//
// Транзакция объединяет обновление попытки и вставку новых решений: иначе
// возможна попытка с применённым результатом, но без записи о том, каким
// выбором он получен.
func (r *AttemptRepository) Update(ctx context.Context, updated attempt.Attempt) (attempt.Attempt, error) {
	if err := ctx.Err(); err != nil {
		return attempt.Attempt{}, err
	}

	if err := validateForPersistence(updated); err != nil {
		return attempt.Attempt{}, err
	}

	row, err := attemptRowOf(updated)
	if err != nil {
		return attempt.Attempt{}, err
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	// Копия делается до транзакции: входной агрегат менять нельзя, а версию
	// вернуть нужно уже новую.
	saved := updated.Clone()

	err = r.inTransaction(queryCtx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		nextVersion, err := r.applyOptimisticUpdate(queryCtx, tx, row)
		if err != nil {
			return err
		}

		persisted, err := r.persistedDecisionCount(queryCtx, tx, updated)
		if err != nil {
			return err
		}

		if err := r.insertDecisions(queryCtx, tx, updated.ID, updated.Decisions[persisted:], persisted); err != nil {
			return err
		}

		saved.Version = int(nextVersion)

		return nil
	})
	if err != nil {
		return attempt.Attempt{}, err
	}

	return saved, nil
}

// applyOptimisticUpdate обновляет строку попытки при совпадении версии.
//
// Ноль обновлённых строк означает либо отсутствие попытки, либо устаревшую
// версию. Различие важно: in-memory адаптер задаёт этот контракт, и сервис
// по-разному реагирует на «не найдено» и «изменено другим запросом».
func (r *AttemptRepository) applyOptimisticUpdate(
	ctx context.Context,
	tx pgx.Tx,
	row attemptRow,
) (int64, error) {
	var nextVersion int64

	err := tx.QueryRow(ctx, updateAttemptQuery,
		row.ID, row.CurrentNodeID, row.Status, row.Score, row.Outcome, row.UpdatedAt,
		row.CompletedAt, row.RevealedNodeIDs, row.AppliedSkillEffects, row.Version,
	).Scan(&nextVersion)

	if err == nil {
		return nextVersion, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("обновить попытку: %w", MapError(err))
	}

	var exists int
	if err := tx.QueryRow(ctx, selectAttemptExistsQuery, row.ID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, attempt.ErrNotFound
		}

		return 0, fmt.Errorf("проверить существование попытки: %w", MapError(err))
	}

	return 0, attempt.ErrConcurrentUpdate
}

// persistedDecisionCount возвращает количество уже сохранённых решений и
// проверяет, что переданная история продолжает сохранённую, а не заменяет её.
func (r *AttemptRepository) persistedDecisionCount(
	ctx context.Context,
	tx pgx.Tx,
	updated attempt.Attempt,
) (int, error) {
	var persisted int

	if err := tx.QueryRow(ctx, countDecisionsQuery, string(updated.ID)).Scan(&persisted); err != nil {
		return 0, fmt.Errorf("посчитать сохранённые решения: %w", MapError(err))
	}

	if persisted > len(updated.Decisions) {
		return 0, fmt.Errorf("%w: сохранено %d решений, передано %d",
			ErrHistoryRewrite, persisted, len(updated.Decisions))
	}

	if persisted == 0 {
		return 0, nil
	}

	var lastKey string
	if err := tx.QueryRow(ctx, selectLastDecisionKeyQuery, string(updated.ID)).Scan(&lastKey); err != nil {
		return 0, fmt.Errorf("прочитать последнее сохранённое решение: %w", MapError(err))
	}

	if incoming := string(updated.Decisions[persisted-1].IdempotencyKey); incoming != lastKey {
		return 0, fmt.Errorf("%w: решение %d сохранено с ключом %q, передан %q",
			ErrHistoryRewrite, persisted, lastKey, incoming)
	}

	return persisted, nil
}

// insertDecisions добавляет решения, начиная с указанного смещения.
func (r *AttemptRepository) insertDecisions(
	ctx context.Context,
	tx pgx.Tx,
	attemptID attempt.ID,
	decisions []attempt.Decision,
	offset int,
) error {
	for i, decision := range decisions {
		// Порядковые номера начинаются с единицы: нулевой ordinal запрещён
		// ограничением схемы.
		row, err := decisionRowOf(decision, offset+i+1)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, insertDecisionQuery,
			string(attemptID), row.Ordinal, row.NodeID, row.ChoiceID, row.ChoiceLabel,
			row.IdempotencyKey, row.Consequence, row.Criticality, row.RiskTags,
			row.SkillEffects, row.ScoreDelta, row.CreatedAt, row.RevealedNodeIDs,
			row.ResultingNodeID, row.Completed, row.Outcome, row.ScoreAfter)
		if err != nil {
			return fmt.Errorf("сохранить решение %d: %w", row.Ordinal, MapError(err))
		}
	}

	return nil
}

func (r *AttemptRepository) loadAttempt(ctx context.Context, tx pgx.Tx, id attempt.ID) (attempt.Attempt, error) {
	var row attemptRow

	err := tx.QueryRow(ctx, selectAttemptQuery, string(id)).Scan(
		&row.ID, &row.ProfileID, &row.ScenarioID, &row.ScenarioVersion, &row.CurrentNodeID,
		&row.Status, &row.Score, &row.Outcome, &row.StartedAt, &row.UpdatedAt, &row.CompletedAt,
		&row.RevealedNodeIDs, &row.AppliedSkillEffects, &row.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attempt.Attempt{}, attempt.ErrNotFound
		}

		return attempt.Attempt{}, fmt.Errorf("прочитать попытку: %w", MapError(err))
	}

	decisions, err := r.loadDecisions(ctx, tx, id)
	if err != nil {
		return attempt.Attempt{}, err
	}

	return attemptOf(row, decisions)
}

func (r *AttemptRepository) loadDecisions(ctx context.Context, tx pgx.Tx, id attempt.ID) ([]attempt.Decision, error) {
	rows, err := tx.Query(ctx, selectDecisionsQuery, string(id))
	if err != nil {
		return nil, fmt.Errorf("прочитать решения попытки: %w", MapError(err))
	}

	defer rows.Close()

	decisions := []attempt.Decision{}

	for rows.Next() {
		var row decisionRow

		err := rows.Scan(
			&row.Ordinal, &row.NodeID, &row.ChoiceID, &row.ChoiceLabel, &row.IdempotencyKey,
			&row.Consequence, &row.Criticality, &row.RiskTags, &row.SkillEffects,
			&row.ScoreDelta, &row.CreatedAt, &row.RevealedNodeIDs, &row.ResultingNodeID,
			&row.Completed, &row.Outcome, &row.ScoreAfter)
		if err != nil {
			return nil, fmt.Errorf("прочитать решение попытки: %w", MapError(err))
		}

		decision, err := decisionOf(row)
		if err != nil {
			return nil, err
		}

		decisions = append(decisions, decision)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать решения попытки: %w", MapError(err))
	}

	return decisions, nil
}

// inTransaction выполняет работу в транзакции и откатывает её при любой ошибке.
func (r *AttemptRepository) inTransaction(
	ctx context.Context,
	isolation pgx.TxIsoLevel,
	work func(tx pgx.Tx) error,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolation})
	if err != nil {
		return fmt.Errorf("начать транзакцию: %w", MapError(err))
	}

	if err := work(tx); err != nil {
		// Ошибка отката не заменяет исходную причину: она интересна только
		// в логах, а решение принимается по первой ошибке.
		_ = tx.Rollback(ctx)

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("зафиксировать транзакцию: %w", MapError(err))
	}

	return nil
}

// validateForPersistence отвергает агрегат, который заведомо нарушит схему.
//
// Проверка дублирует ограничения базы намеренно: сообщение об ошибке остаётся
// доменным, а не превращается в текст нарушенного constraint.
func validateForPersistence(source attempt.Attempt) error {
	switch {
	case source.ID == "":
		return attempt.ErrEmptyAttemptID
	case source.ProfileID == "":
		return attempt.ErrEmptyProfileID
	case source.ScenarioID == "":
		return attempt.ErrEmptyScenarioID
	case source.ScenarioVersion < 1:
		return attempt.ErrInvalidScenarioVersion
	case source.CurrentNodeID == "":
		return attempt.ErrEmptyNode
	case source.StartedAt.IsZero():
		return attempt.ErrEmptyStartTime
	case len(source.RevealedNodeIDs) == 0:
		return attempt.ErrNothingRevealed
	}

	if source.Version < 1 {
		return fmt.Errorf("%w: версия попытки должна быть положительной, получено %d",
			ErrIntegrityViolation, source.Version)
	}

	if source.Score < attempt.MinScore || source.Score > attempt.MaxScore {
		return fmt.Errorf("%w: результат попытки вне диапазона %d..%d, получено %d",
			ErrIntegrityViolation, attempt.MinScore, attempt.MaxScore, source.Score)
	}

	return nil
}
