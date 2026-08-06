package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/google/uuid"
)

// User 是 main 包中的用户结构体，用于控制 JSON 序列化的 key 名称
// （database.User 由 sqlc 生成，没有 JSON tag，直接序列化会输出大写驼峰 key）
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// databaseUser 将 database.User 映射为 main 包的 User
func databaseUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
}

// 创建用户
func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//decode request
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create user
	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userJson, err := json.Marshal(databaseUser(dbUser))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(userJson)
}

// 重置用户
func (cfg *apiConfig) resetUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	//platform check
	platform := cfg.platform
	if platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	//delete users
	err := cfg.dbQueries.DeleteUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Delete users failed"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users reset"))
}
