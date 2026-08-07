package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/auth"
	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/Ghost-Unison/Chirpy/internal/platform"
)

// Handler 持有用户相关处理器所需的依赖。
type Handler struct {
	db        *database.Queries
	secretKey string
}

// NewHandler 构造一个持有 db 与 JWT 密钥的 users.Handler。
func NewHandler(db *database.Queries, secretKey string) *Handler {
	return &Handler{db: db, secretKey: secretKey}
}

// Create 处理 POST /api/users: 创建用户。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//decode request
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// validate request: email and password must not be empty
	if req.Email == "" || req.Password == "" {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Email and password are required"})
		return
	}

	// create user
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Hashing password failed"})
		return
	}
	dbUser, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Creating user failed"})
		return
	}

	platform.WriteJSON(w, http.StatusOK, databaseUser(dbUser))
}

// Login 处理 POST /api/login: 校验邮箱密码后签发 access_token 与 refresh_token。
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//decode request
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// validate request: email and password must not be empty
	if req.Email == "" || req.Password == "" {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Email and password are required"})
		return
	}

	//get user by email
	dbUser, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Incorrect email or password"})
		return
	}

	// check password
	match, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !match {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Incorrect email or password"})
		return
	}

	// generate access_token
	token, err := auth.MakeJWT(dbUser.ID, h.secretKey)
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Failed to create token"})
		return
	}
	user := databaseUser(dbUser)
	user.Token = token

	// generate refresh_token
	refreshToken := auth.MakeRefreshToken()
	user.RefreshToken = refreshToken

	// store refresh_token
	_, err = h.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		UserID:    dbUser.ID,
	})
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Failed to save refresh token"})
		return
	}

	platform.WriteJSON(w, http.StatusOK, user)
}

// Update 处理 PUT /api/users: 鉴权后更新用户 email 或 password。
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//jwt auth
	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Invalid token"})
		return
	}

	userId, err := auth.ValidateJWT(jwt, h.secretKey)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Invalid token"})
		return
	}

	//decode request
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	if req.Email == "" || req.Password == "" {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Email or password is required"})
		return
	}

	//Argon2id 每次哈希使用随机盐，所以 HashPassword(相同密码) 每次产生的哈希都不同
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Hashing password failed"})
		return
	}

	// update user
	newDbUser, err := h.db.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		ID:             userId,
	})
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Failed to update user"})
		return
	}

	platform.WriteJSON(w, http.StatusOK, databaseUser(newDbUser))
}
