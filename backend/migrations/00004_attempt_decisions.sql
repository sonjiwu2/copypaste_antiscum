-- +goose Up
-- +goose StatementBegin

-- Решение пользователя вместе со снимком рассчитанного сервером перехода.
--
-- Снимки choice_label, consequence, risk_tags, skill_effects и score_after
-- обязательны: повтор запроса с тем же ключом должен вернуть тот же ответ
-- даже после того, как текст сценария изменится. Одних идентификаторов для
-- этого недостаточно.
CREATE TABLE attempt_decisions (
    attempt_id        text        NOT NULL,
    -- Порядковый номер решения внутри попытки, начиная с единицы.
    ordinal           integer     NOT NULL CHECK (ordinal >= 1),
    node_id           text        NOT NULL CHECK (node_id <> ''),
    choice_id         text        NOT NULL CHECK (choice_id <> ''),
    choice_label      text        NOT NULL,
    idempotency_key   text        NOT NULL CHECK (idempotency_key <> ''),
    consequence       jsonb       NOT NULL,
    criticality       text        NOT NULL CHECK (criticality IN ('low', 'medium', 'high')),
    risk_tags         text[]      NOT NULL DEFAULT '{}',
    skill_effects     jsonb       NOT NULL DEFAULT '[]',
    score_delta       integer     NOT NULL,
    created_at        timestamptz NOT NULL,
    revealed_node_ids text[]      NOT NULL DEFAULT '{}',
    resulting_node_id text        NOT NULL CHECK (resulting_node_id <> ''),
    completed         boolean     NOT NULL,
    outcome           text        NULL CHECK (outcome IS NULL OR outcome IN ('safe', 'unsafe')),
    score_after       integer     NOT NULL CHECK (score_after BETWEEN 0 AND 100),

    CONSTRAINT attempt_decisions_pkey PRIMARY KEY (attempt_id, ordinal),

    -- Второй рубеж идемпотентности после проверки в сервисе.
    CONSTRAINT attempt_decisions_idempotency_key UNIQUE (attempt_id, idempotency_key),

    CONSTRAINT attempt_decisions_attempt_id_fkey FOREIGN KEY (attempt_id)
        REFERENCES attempts (id) ON DELETE CASCADE,

    CONSTRAINT attempt_decisions_completion_consistent CHECK (
        (completed = false AND outcome IS NULL)
        OR
        (completed = true AND outcome IS NOT NULL)
    )
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE attempt_decisions;

-- +goose StatementEnd
