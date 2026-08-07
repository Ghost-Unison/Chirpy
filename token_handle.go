package main

import (
	"net/http"
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/auth"
)

/*
不需要请求体 需要Header中的refresh_token.

在数据库中查找此refresh_token，不存在或已经过期则返回401
否则返回200和新的access_token
*/
func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {

	//get refresh token from header
	refreshTokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Invalid refresh token"})
		return
	}

	//look up refresh token in database，不存在或出错都返回401
	refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshTokenStr)
	//查不到也会返回错误 sql.ErrNoRows
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Invalid refresh token"})
		return
	}
	//判断sql.NullTime是否为空 用 valid
	if refreshToken.RevokedAt.Valid {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Invalid refresh token"})
		return
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Expired refresh token"})
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Failed to generate access token"})
		return
	}
	writeJSON(w, http.StatusOK,
		struct {
			Token string `json:"token"`
		}{Token: accessToken})
}

/*
不需要请求体 需要Header中的refresh_token。

根据Header中的refresh_token，在数据库中撤销对应的记录（设置revoked_at为当前时间戳）。
无论token是否存在、是否已撤销，都返回204。
*/
func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	//get refresh token from header
	refreshTokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Invalid refresh token"})
		return
	}

	//revoke the refresh token in database
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshTokenStr)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Failed to revoke refresh token"})
		return
	}
	//return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
