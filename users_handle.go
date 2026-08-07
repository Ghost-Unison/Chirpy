package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/auth"
	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/google/uuid"
)

// 创建用户、用户登录通用请求体
type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User 是 main 包中的用户结构体，用于控制 JSON 序列化的 key 名称
// （database.User 由 sqlc 生成，没有 JSON tag，直接序列化会输出大写驼峰 key）
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Invalid request payload"})
		return
	}

	// validate request: email and password must not be empty
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Email and password are required"})
		return
	}

	// create user
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Hashing password failed"})
		return
	}
	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Creating user failed"})
		return
	}

	writeJSON(w, http.StatusOK, databaseUser(dbUser))
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

// 用户登录
func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//decode request
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Invalid request payload"})
		return
	}

	// validate request: email and password must not be empty
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Email and password are required"})
		return
	}

	//get user by email
	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Incorrect email or password"})
		return
	}

	// check password
	match, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !match {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Incorrect email or password"})
		return
	}

	// generate access_token
	token, err := auth.MakeJWT(dbUser.ID, cfg.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Failed to create token"})
		return
	}
	user := databaseUser(dbUser)
	user.Token = token

	// generate refresh_token
	refreshToken := auth.MakeRefreshToken()
	user.RefreshToken = refreshToken

	// store refresh_token
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		UserID:    dbUser.ID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Failed to save refresh token"})
		return
	}

	writeJSON(w, http.StatusOK, user)

}
