package auth

import (
	"net/http"
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/Ghost-Unison/Chirpy/internal/platform"
)

// Handler 持有 refresh/revoke 处理器所需的依赖。
// auth 包同时承载无状态工具函数（jwt/password/refresh_token/webhook）
// 与这里基于数据库与密钥的有状态处理器。
type Handler struct {
	db        *database.Queries
	secretKey string
}

// NewHandler 构造一个持有 db 与 JWT 密钥的 auth.Handler。
func NewHandler(db *database.Queries, secretKey string) *Handler {
	return &Handler{db: db, secretKey: secretKey}
}

/*
不需要请求体 需要Header中的refresh_token.

在数据库中查找此refresh_token，不存在或已经过期则返回401
否则返回200和新的access_token
*/
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	//get refresh token from header
	refreshTokenStr, err := GetBearerToken(r.Header)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	//look up refresh token in database，不存在或出错都返回401
	refreshToken, err := h.db.GetRefreshToken(r.Context(), refreshTokenStr)
	//查不到也会返回错误 sql.ErrNoRows
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Invalid refresh token"})
		return
	}
	//判断sql.NullTime是否为空 用 valid
	if refreshToken.RevokedAt.Valid {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Invalid refresh token"})
		return
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Expired refresh token"})
		return
	}

	accessToken, err := MakeJWT(refreshToken.UserID, h.secretKey)
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Failed to generate access token"})
		return
	}
	platform.WriteJSON(w, http.StatusOK,
		struct {
			Token string `json:"token"`
		}{Token: accessToken})
}

/*
不需要请求体 需要Header中的refresh_token。

根据Header中的refresh_token，在数据库中撤销对应的记录（设置revoked_at为当前时间戳）。
无论token是否存在、是否已撤销，都返回204。
*/
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	//get refresh token from header
	refreshTokenStr, err := GetBearerToken(r.Header)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	//revoke the refresh token in database
	err = h.db.RevokeRefreshToken(r.Context(), refreshTokenStr)
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Failed to revoke refresh token"})
		return
	}
	//return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
