-- +goose Up
-- +goose StatementBegin

ALTER TABLE weekly_tests
    DROP CONSTRAINT weekly_tests_source_check;

ALTER TABLE weekly_tests
    ADD CONSTRAINT weekly_tests_source_check
    CHECK (source IN ('grok', 'groq', 'fallback'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE weekly_tests
SET source = 'fallback', model = ''
WHERE source = 'groq';

ALTER TABLE weekly_tests
    DROP CONSTRAINT weekly_tests_source_check;

ALTER TABLE weekly_tests
    ADD CONSTRAINT weekly_tests_source_check
    CHECK (source IN ('grok', 'fallback'));

-- +goose StatementEnd
