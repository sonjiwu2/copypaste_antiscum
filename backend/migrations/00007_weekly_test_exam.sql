-- +goose Up
-- +goose StatementBegin

ALTER TABLE weekly_tests
    DROP CONSTRAINT weekly_tests_submission_consistent,
    DROP CONSTRAINT weekly_tests_correct_count_check,
    DROP CONSTRAINT weekly_tests_total_count_check,
    DROP CONSTRAINT weekly_tests_earned_xp_check;

ALTER TABLE weekly_tests
    ADD COLUMN passed boolean NULL,
    ADD COLUMN timed_out boolean NULL,
    ADD COLUMN lives_left integer NULL;

-- Старые завершённые тесты из 5 вопросов сохраняются. Новые поля для них
-- восстанавливаются из уже записанного результата.
UPDATE weekly_tests
SET passed = (score >= 90),
    timed_out = false,
    lives_left = GREATEST(0, 3 - (total_count - correct_count))
WHERE completed_at IS NOT NULL;

ALTER TABLE weekly_tests
    ADD CONSTRAINT weekly_tests_correct_count_check CHECK (correct_count BETWEEN 0 AND 20),
    ADD CONSTRAINT weekly_tests_total_count_check CHECK (total_count IN (5, 20)),
    ADD CONSTRAINT weekly_tests_earned_xp_check CHECK (earned_xp BETWEEN 0 AND 100),
    ADD CONSTRAINT weekly_tests_lives_left_check CHECK (lives_left BETWEEN 0 AND 3),
    ADD CONSTRAINT weekly_tests_submission_consistent CHECK (
        (completed_at IS NULL AND answers IS NULL AND correct_count IS NULL
            AND total_count IS NULL AND score IS NULL AND earned_xp IS NULL
            AND passed IS NULL AND timed_out IS NULL AND lives_left IS NULL)
        OR
        (completed_at IS NOT NULL AND answers IS NOT NULL AND correct_count IS NOT NULL
            AND total_count IS NOT NULL AND score IS NOT NULL AND earned_xp IS NOT NULL
            AND passed IS NOT NULL AND timed_out IS NOT NULL AND lives_left IS NOT NULL)
    ),
    ADD CONSTRAINT weekly_tests_result_rules CHECK (
        completed_at IS NULL
        OR total_count = 5
        OR (passed = true AND timed_out = false AND correct_count >= 18 AND earned_xp = 100)
        OR (passed = false AND earned_xp = 0)
    );

-- +goose StatementEnd

-- +goose Down
-- Миграция намеренно forward-only: удаление новых полей уничтожило бы данные
-- уже пройденных экзаменов. История сохраняется даже при откате версии кода.
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
