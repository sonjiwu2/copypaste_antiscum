-- +goose Up
-- +goose StatementBegin

CREATE TABLE weekly_tests (
    id            text        NOT NULL CHECK (id <> ''),
    profile_id    text        NOT NULL,
    week_start    date        NOT NULL,
    title         text        NOT NULL CHECK (title <> ''),
    intro         text        NOT NULL CHECK (intro <> ''),
    source        text        NOT NULL CHECK (source IN ('grok', 'fallback')),
    model         text        NOT NULL DEFAULT '',
    questions     jsonb       NOT NULL,
    answers       jsonb       NULL,
    correct_count integer     NULL CHECK (correct_count BETWEEN 0 AND 5),
    total_count   integer     NULL CHECK (total_count = 5),
    score         integer     NULL CHECK (score BETWEEN 0 AND 100),
    earned_xp     integer     NULL CHECK (earned_xp BETWEEN 0 AND 25),
    completed_at  timestamptz NULL,
    created_at    timestamptz NOT NULL,

    CONSTRAINT weekly_tests_pkey PRIMARY KEY (id),
    CONSTRAINT weekly_tests_profile_id_fkey FOREIGN KEY (profile_id)
        REFERENCES profiles (id),
    CONSTRAINT weekly_tests_profile_week_key UNIQUE (profile_id, week_start),
    CONSTRAINT weekly_tests_submission_consistent CHECK (
        (completed_at IS NULL AND answers IS NULL AND correct_count IS NULL
            AND total_count IS NULL AND score IS NULL AND earned_xp IS NULL)
        OR
        (completed_at IS NOT NULL AND answers IS NOT NULL AND correct_count IS NOT NULL
            AND total_count IS NOT NULL AND score IS NOT NULL AND earned_xp IS NOT NULL)
    )
);

CREATE INDEX weekly_tests_profile_created_idx
    ON weekly_tests (profile_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE weekly_tests;

-- +goose StatementEnd

