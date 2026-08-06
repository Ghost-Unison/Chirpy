-- +goose Up
-- 创建 users 表，存储用户基本信息
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    email text NOT NULL UNIQUE
);

-- +goose Down
-- 回滚：删除 users 表
DROP TABLE users;