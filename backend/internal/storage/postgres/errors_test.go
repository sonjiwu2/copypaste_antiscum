package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

// Разбор ошибок идёт по кодам и именам ограничений, а не по тексту сообщения:
// текст зависит от локали и версии сервера.
func TestMapError(t *testing.T) {
	testCases := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "без ошибки",
			err:     nil,
			wantErr: nil,
		},
		{
			name:    "отсутствующая строка",
			err:     pgx.ErrNoRows,
			wantErr: attempt.ErrNotFound,
		},
		{
			name: "повторный идентификатор попытки",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: postgres.ConstraintAttemptsPrimaryKey,
			},
			wantErr: attempt.ErrDuplicateAttempt,
		},
		{
			// Параллельный запрос уже записал этот шаг: сервис обязан
			// перечитать попытку, а не сообщать пользователю об ошибке.
			name: "повторный ключ идемпотентности",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: postgres.ConstraintDecisionIdempotencyKey,
			},
			wantErr: attempt.ErrConcurrentUpdate,
		},
		{
			name: "повторный порядковый номер решения",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: postgres.ConstraintDecisionsPrimaryKey,
			},
			wantErr: attempt.ErrConcurrentUpdate,
		},
		{
			name:    "сбой сериализации транзакции",
			err:     &pgconn.PgError{Code: "40001"},
			wantErr: attempt.ErrConcurrentUpdate,
		},
		{
			name:    "взаимная блокировка",
			err:     &pgconn.PgError{Code: "40P01"},
			wantErr: attempt.ErrConcurrentUpdate,
		},
		{
			name: "нарушение внешнего ключа профиля",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: postgres.ConstraintAttemptProfileForeignKey,
			},
			wantErr: postgres.ErrIntegrityViolation,
		},
		{
			name:    "нарушение проверочного ограничения",
			err:     &pgconn.PgError{Code: "23514", ConstraintName: "attempts_completion_consistent"},
			wantErr: postgres.ErrIntegrityViolation,
		},
		{
			name:    "отмена вызывающего важнее ошибки драйвера",
			err:     fmt.Errorf("запрос прерван: %w", context.Canceled),
			wantErr: context.Canceled,
		},
		{
			name:    "истёкший дедлайн",
			err:     fmt.Errorf("запрос прерван: %w", context.DeadlineExceeded),
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mapped := postgres.MapError(testCase.err)

			if testCase.wantErr == nil {
				if mapped != nil {
					t.Fatalf("ошибка = %v, ожидалась nil", mapped)
				}

				return
			}

			if !errors.Is(mapped, testCase.wantErr) {
				t.Fatalf("ошибка = %v, ожидалась %v", mapped, testCase.wantErr)
			}
		})
	}
}

// Неизвестная ошибка соединения не должна превращаться в доменную:
// иначе обрыв связи выглядел бы как отсутствующая попытка.
func TestMapErrorKeepsUnknownErrors(t *testing.T) {
	original := errors.New("connection refused")

	mapped := postgres.MapError(original)

	if !errors.Is(mapped, original) {
		t.Fatalf("ошибка = %v, ожидалась исходная", mapped)
	}

	for _, domainErr := range []error{
		attempt.ErrNotFound,
		attempt.ErrDuplicateAttempt,
		attempt.ErrConcurrentUpdate,
		postgres.ErrIntegrityViolation,
	} {
		if errors.Is(mapped, domainErr) {
			t.Errorf("неизвестная ошибка распознана как %v", domainErr)
		}
	}
}

func TestIsIntegrityViolation(t *testing.T) {
	mapped := postgres.MapError(&pgconn.PgError{Code: "23502", ConstraintName: "attempts_score_check"})

	if !postgres.IsIntegrityViolation(mapped) {
		t.Errorf("нарушение целостности не распознано: %v", mapped)
	}

	if postgres.IsIntegrityViolation(attempt.ErrNotFound) {
		t.Error("доменная ошибка не должна считаться нарушением целостности")
	}
}
