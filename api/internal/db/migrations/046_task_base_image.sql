-- +goose Up
ALTER TABLE tasks ADD COLUMN base_image TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tasks DROP COLUMN base_image;
