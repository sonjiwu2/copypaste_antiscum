package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
)

// Имена ограничений схемы. Они заданы явно в миграциях, потому что различить
// нарушения можно только по имени: код 23505 одинаков для всех уникальных
// индексов.
const (
	ConstraintAttemptsPrimaryKey       = "attempts_pkey"
	ConstraintDecisionIdempotencyKey   = "attempt_decisions_idempotency_key"
	ConstraintDecisionsPrimaryKey      = "attempt_decisions_pkey"
	ConstraintAttemptProfileForeignKey = "attempts_profile_id_fkey"
	ConstraintAttemptScenarioVersion   = "attempts_scenario_version_fkey"
)

// Коды ошибок PostgreSQL, которые адаптер разбирает осознанно.
const (
	codeUniqueViolation      = "23505"
	codeForeignKeyViolation  = "23503"
	codeNotNullViolation     = "23502"
	codeCheckViolation       = "23514"
	codeSerializationFailure = "40001"
	codeDeadlockDetected     = "40P01"
)

// ErrIntegrityViolation — база отвергла данные, нарушающие инвариант схемы.
//
// Это дефект приложения, а не действие пользователя: наружу такая ошибка
// уходит как внутренняя ошибка сервера.
var ErrIntegrityViolation = errors.New("нарушение целостности данных")

// MapError переводит ошибку драйвера в доменную или в понятную инфраструктурную.
//
// Это единственное место разбора кодов PostgreSQL: сравнение строк ошибок
// запрещено, а сырые сообщения драйвера не должны доходить до клиента.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	// Отмена вызывающего важнее любой ошибки драйвера: клиент уже ушёл.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return attempt.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case codeUniqueViolation:
		return mapUniqueViolation(pgErr)
	case codeSerializationFailure, codeDeadlockDetected:
		// Конкурентный переход, а не поломка: сервис перечитает попытку.
		return attempt.ErrConcurrentUpdate
	case codeForeignKeyViolation, codeNotNullViolation, codeCheckViolation:
		return fmt.Errorf("%w: %s", ErrIntegrityViolation, pgErr.ConstraintName)
	default:
		return err
	}
}

func mapUniqueViolation(pgErr *pgconn.PgError) error {
	switch pgErr.ConstraintName {
	case ConstraintAttemptsPrimaryKey:
		return attempt.ErrDuplicateAttempt
	case ConstraintDecisionIdempotencyKey, ConstraintDecisionsPrimaryKey:
		// Второй рубеж идемпотентности: параллельный запрос уже записал этот
		// шаг. Сервис разрешит ситуацию повторным чтением попытки.
		return attempt.ErrConcurrentUpdate
	default:
		return fmt.Errorf("%w: %s", ErrIntegrityViolation, pgErr.ConstraintName)
	}
}

// IsIntegrityViolation сообщает, отвергла ли база данные по ограничению схемы.
func IsIntegrityViolation(err error) bool {
	return errors.Is(err, ErrIntegrityViolation)
}
