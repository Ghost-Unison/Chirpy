-- +goose Up
-- 创建 refresh_tokens 表，存储用户刷新令牌
CREATE TABLE refresh_tokens (
    token text NOT NULL PRIMARY KEY,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamp DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 day' NOT NULL,
    revoked_at timestamp DEFAULT NULL
);

-- +goose Down
-- 回滚：删除 refresh_tokens 表
DROP TABLE refresh_tokens;
