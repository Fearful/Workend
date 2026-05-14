-- +goose Up
ALTER TABLE tasks ADD COLUMN network_mode TEXT NOT NULL DEFAULT 'default';
ALTER TABLE tasks ADD COLUMN exposed_ports TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE tasks DROP COLUMN exposed_ports;
ALTER TABLE tasks DROP COLUMN network_mode;
