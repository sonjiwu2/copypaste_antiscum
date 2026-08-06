-- +goose Up
-- +goose StatementBegin

-- Архив версий сценариев. Содержимое остаётся собственностью JSON-файлов:
-- таблица только хранит уже проверенные версии, чтобы начатая попытка
-- доигрывалась на своей версии даже после выпуска новой.
--
-- content_hash защищает от правки содержимого без поднятия version:
-- та же пара (scenario_id, version) с другим хэшем — ошибка запуска.
CREATE TABLE scenario_versions (
    scenario_id       text        NOT NULL CHECK (scenario_id <> ''),
    version           integer     NOT NULL CHECK (version >= 1),
    slug              text        NOT NULL,
    role              text        NOT NULL CHECK (role IN ('buyer', 'seller')),
    title             text        NOT NULL CHECK (title <> ''),
    description       text        NOT NULL,
    difficulty        text        NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    -- Валидатор сценариев не требует положительной длительности, поэтому
    -- ограничение здесь не строже правил контента.
    estimated_minutes integer     NOT NULL CHECK (estimated_minutes >= 0),
    is_active         boolean     NOT NULL,
    content           jsonb       NOT NULL,
    content_hash      text        NOT NULL CHECK (content_hash <> ''),
    created_at        timestamptz NOT NULL,

    CONSTRAINT scenario_versions_pkey PRIMARY KEY (scenario_id, version)
);

-- Каталог показывает только активные версии сценариев.
CREATE INDEX scenario_versions_active_idx
    ON scenario_versions (scenario_id, version DESC)
    WHERE is_active;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE scenario_versions;

-- +goose StatementEnd
