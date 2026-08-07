-- +goose Up
-- 给 users 表添加 is_chirpy_red 字段
ALTER TABLE users ADD COLUMN is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
-- 回滚：删除 is_chirpy_red 字段
ALTER TABLE users DROP COLUMN is_chirpy_red;
