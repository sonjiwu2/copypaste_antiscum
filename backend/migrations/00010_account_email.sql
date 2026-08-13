-- +goose Up
-- +goose StatementBegin

ALTER TABLE accounts RENAME COLUMN username TO email;
ALTER TABLE accounts RENAME COLUMN username_normalized TO email_normalized;
ALTER TABLE accounts RENAME CONSTRAINT accounts_username_normalized_key TO accounts_email_normalized_key;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE accounts RENAME CONSTRAINT accounts_email_normalized_key TO accounts_username_normalized_key;
ALTER TABLE accounts RENAME COLUMN email_normalized TO username_normalized;
ALTER TABLE accounts RENAME COLUMN email TO username;

-- +goose StatementEnd
