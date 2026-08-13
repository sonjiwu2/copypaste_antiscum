-- +goose Up
-- +goose StatementBegin

-- Учётная запись отделена от игрового профиля: профиль остаётся владельцем
-- попыток, а аккаунт добавляет вход с нескольких устройств. Один профиль
-- может принадлежать только одному аккаунту.
CREATE TABLE accounts (
    id                  text        PRIMARY KEY CHECK (id <> ''),
    profile_id          text        NOT NULL UNIQUE,
    username            text        NOT NULL,
    username_normalized text        NOT NULL,
    password_hash       text        NOT NULL CHECK (password_hash <> ''),
    created_at          timestamptz NOT NULL,
    updated_at          timestamptz NOT NULL,

    CONSTRAINT accounts_profile_id_fkey FOREIGN KEY (profile_id)
        REFERENCES profiles (id) ON DELETE CASCADE,
    CONSTRAINT accounts_username_normalized_key UNIQUE (username_normalized)
);

-- В cookie хранится случайный токен, а в базе только SHA-256 этого токена.
-- Компрометация таблицы сессий поэтому не даёт готовых bearer-токенов.
CREATE TABLE auth_sessions (
    token_hash   text        PRIMARY KEY CHECK (char_length(token_hash) = 64),
    account_id   text        NOT NULL,
    created_at   timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    expires_at   timestamptz NOT NULL CHECK (expires_at > created_at),

    CONSTRAINT auth_sessions_account_id_fkey FOREIGN KEY (account_id)
        REFERENCES accounts (id) ON DELETE CASCADE
);

CREATE INDEX auth_sessions_account_idx
    ON auth_sessions (account_id, expires_at DESC);
CREATE INDEX auth_sessions_expiry_idx
    ON auth_sessions (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE auth_sessions;
DROP TABLE accounts;

-- +goose StatementEnd
