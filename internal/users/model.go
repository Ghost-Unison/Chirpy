// Package users 负责用户资源的 HTTP 处理与模型映射: 创建、登录、更新。
package users

import (
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/google/uuid"
)

// userRequest 是创建用户、用户登录、更新用户通用的请求体
type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User 是面向 API 输出的用户结构体，用于控制 JSON 序列化的 key 名称
// （database.User 由 sqlc 生成，没有 JSON tag，直接序列化会输出大写驼峰 key）
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

// databaseUser 将 database.User 映射为 users.User
func databaseUser(dbUser database.User) User {
	return User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}
}
