-- +goose Up
-- +goose StatementBegin

-- Одно прохождение зафиксированной версии сценария.
--
-- version — оптимистичная блокировка: обновление проходит только при
-- совпадении версии, поэтому два параллельных выбора не применяются оба.
CREATE TABLE attempts (
    id                    text        NOT NULL CHECK (id <> ''),
    profile_id            text        NOT NULL,
    scenario_id           text        NOT NULL,
    scenario_version      integer     NOT NULL CHECK (scenario_version >= 1),
    current_node_id       text        NOT NULL CHECK (current_node_id <> ''),
    status                text        NOT NULL CHECK (status IN ('in_progress', 'completed')),
    score                 integer     NOT NULL CHECK (score BETWEEN 0 AND 100),
    outcome               text        NULL CHECK (outcome IS NULL OR outcome IN ('safe', 'unsafe')),
    started_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    completed_at          timestamptz NULL,
    -- Порядок раскрытия узлов: по нему восстанавливается переписка.
    -- Попытка начинается минимум с одного раскрытого узла.
    revealed_node_ids     text[]      NOT NULL DEFAULT '{}'
                                      CHECK (cardinality(revealed_node_ids) >= 1),
    applied_skill_effects jsonb       NOT NULL DEFAULT '{}',
    version               bigint      NOT NULL CHECK (version >= 1),

    CONSTRAINT attempts_pkey PRIMARY KEY (id),

    CONSTRAINT attempts_profile_id_fkey FOREIGN KEY (profile_id)
        REFERENCES profiles (id),

    -- Версия сценария обязана существовать в архиве, иначе попытку
    -- невозможно будет доиграть после выпуска нового контента.
    CONSTRAINT attempts_scenario_version_fkey FOREIGN KEY (scenario_id, scenario_version)
        REFERENCES scenario_versions (scenario_id, version),

    -- Завершённость выражается тремя полями сразу, и они обязаны быть
    -- согласованы: иначе прогресс посчитает незавершённую попытку.
    CONSTRAINT attempts_completion_consistent CHECK (
        (status = 'in_progress' AND completed_at IS NULL AND outcome IS NULL)
        OR
        (status = 'completed' AND completed_at IS NOT NULL AND outcome IS NOT NULL)
    )
);

-- История последних попыток профиля.
CREATE INDEX attempts_profile_updated_idx
    ON attempts (profile_id, updated_at DESC, id DESC);

-- Прогресс по сценариям: лучший, последний и предыдущий результат.
CREATE INDEX attempts_profile_scenario_completed_idx
    ON attempts (profile_id, scenario_id, completed_at DESC, id DESC)
    WHERE status = 'completed';

-- Незавершённые прохождения показываются отдельно от статистики.
CREATE INDEX attempts_profile_active_idx
    ON attempts (profile_id, updated_at DESC, id DESC)
    WHERE status = 'in_progress';

-- Проверка «есть ли попытки на этой версии» при синхронизации сценариев.
CREATE INDEX attempts_scenario_version_idx
    ON attempts (scenario_id, scenario_version);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE attempts;

-- +goose StatementEnd
