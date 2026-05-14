-- +goose Up
ALTER TABLE user_ssh_keys ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE user_ssh_keys ADD COLUMN last_rotated_at TIMESTAMPTZ;
ALTER TABLE user_pat_credentials ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE user_pat_credentials ADD COLUMN last_rotated_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE user_pat_credentials DROP COLUMN last_rotated_at;
ALTER TABLE user_pat_credentials DROP COLUMN expires_at;
ALTER TABLE user_ssh_keys DROP COLUMN last_rotated_at;
ALTER TABLE user_ssh_keys DROP COLUMN expires_at;
