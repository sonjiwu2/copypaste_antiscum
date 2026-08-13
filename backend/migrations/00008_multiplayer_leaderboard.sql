-- +goose Up
-- +goose StatementBegin

-- Публичная личность анонимного профиля. Секретный id из cookie никогда не
-- попадает в лидерборд; наружу отдаются только выбранные имя и аватар.
ALTER TABLE profiles
    ADD COLUMN display_name text NOT NULL DEFAULT 'Игрок'
        CHECK (char_length(btrim(display_name)) BETWEEN 2 AND 24),
    ADD COLUMN avatar text NOT NULL DEFAULT 'profile'
        CHECK (avatar IN ('profile', 'leader-1', 'leader-2', 'leader-3'));

-- Стабильные непустые имена для уже существующих анонимных игроков.
UPDATE profiles
SET display_name = 'Игрок ' || upper(substr(id, 1, 4))
WHERE id <> 'profile_default_demo';

-- Глобальный рейтинг строится по лучшему результату каждого сценария.
CREATE INDEX attempts_leaderboard_completed_idx
    ON attempts (profile_id, scenario_id, score DESC, completed_at DESC)
    WHERE status = 'completed';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX attempts_leaderboard_completed_idx;
ALTER TABLE profiles DROP COLUMN avatar, DROP COLUMN display_name;

-- +goose StatementEnd
