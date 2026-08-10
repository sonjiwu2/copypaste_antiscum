-- +goose Up
-- +goose StatementBegin

-- Анонимный профиль пользователя: стабильный владелец попыток и прогресса.
-- Пароли и OAuth в MVP не хранятся намеренно.
CREATE TABLE profiles (
    id         text        PRIMARY KEY CHECK (id <> ''),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- Служебный профиль остаётся владельцем попыток, созданных до появления
-- анонимной идентификации. Новые попытки к нему не привязываются: их владелец
-- приходит из cookie. Удалять запись нельзя — на неё ссылаются старые строки
-- attempts, и внешний ключ этого не позволит.
INSERT INTO profiles (id, created_at, updated_at)
VALUES ('profile_default_demo', now(), now());

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE profiles;

-- +goose StatementEnd
