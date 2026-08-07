-- +goose Up
-- 给 users 表添加 hashed_password 字段
ALTER TABLE users ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
-- 回滚：删除 hashed_password 字段
ALTER TABLE users DROP COLUMN hashed_password;