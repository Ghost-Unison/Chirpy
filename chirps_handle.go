// json处理 - 数据校验+敏感词处理
package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Ghost-Unison/Chirpy/internal/auth"
	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/google/uuid"
)

// errorResponse 是统一的错误响应结构，格式为 {"error": "..."}
type errorResponse struct {
	Error string `json:"error"`
}

// chirpRequest 是 POST /api/chirps 的请求体
// userId以jwt中为准，不再从请求体中获取
type chirpRequest struct {
	Body string `json:"body"`
	//UserId uuid.UUID `json:"user_id"`
}

// sqlc Chirp -> database.Chirp
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func databaseChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

func databaseChirps(dbChirps []database.Chirp) []Chirp {
	chirps := []Chirp{} // Initialize an empty slice of Chirp  - 这样如果查询结果为空，返回的是空集合，而不是 nil
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, databaseChirp(dbChirp))
	}
	return chirps
}

// 需要过滤的敏感词集合（统一用小写形式存储，匹配时对单词做 ToLower）
var badWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

// cleanBody 将 body 中的敏感词（作为独立单词，大小写不敏感）替换为 ****。
// 带标点的单词（如 "Sharbert!"）不会被命中，原样保留。
func cleanBody(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

// writeJSON marshals payload to JSON and writes it with the given status code.
// 先 marshal 再 WriteHeader，避免重复调用 WriteHeader 导致状态码被忽略。
func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		// marshal 失败几乎不可能发生，兜底返回 500 且不写 body。
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Error: "Invalid token",
		})
		return
	}
	userId, err := auth.ValidateJWT(jwt, cfg.secretKey)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Error: "Invalid token",
		})
		return
	}

	// decode request
	var req chirpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "Invalid request payload",
		})
		return
	}

	// length judge（按 rune 计数，即字符数而非字节数）
	if utf8.RuneCountInString(req.Body) > 140 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "Body length cannot exceed 140 characters",
		})
		return
	}
	/*
		validate user ID -当 json.NewDecoder().Decode() 解析 JSON 时，会自动尝试把 user_id 字段的字符串解析成 UUID，这里只需要排除空值
		if req.UserId == uuid.Nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "User ID is required",
			})
			return
		}
	*/

	//不再返回 cleaned_body，而是存入数据库
	cleanedBody := cleanBody(req.Body)
	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "Could not create chirp",
		})
		return
	}

	// 成功：返回 201 + 完整 chirp 资源
	writeJSON(w, http.StatusCreated, databaseChirp(dbChirp))
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	dbChirps, err := cfg.dbQueries.QueryChirps(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "Could not query chirps",
		})
		return
	}

	writeJSON(w, http.StatusOK, databaseChirps(dbChirps))
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	chirpID := r.PathValue("chirpID")
	validChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "Invalid chirp ID",
		})
		return
	}

	dbChirp, err := cfg.dbQueries.QueryChirp(r.Context(), validChirpID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "Could not query chirp",
		})
		return
	}

	writeJSON(w, http.StatusOK, databaseChirp(dbChirp))
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//jwt auth
	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Invalid token"})
		return
	}
	userId, err := auth.ValidateJWT(jwt, cfg.secretKey)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "Invalid token"})
		return
	}

	//get chirp
	chirpID := r.PathValue("chirpID")
	validChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Invalid chirp ID"})
		return
	}

	dbChirp, err := cfg.dbQueries.QueryChirp(r.Context(), validChirpID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Could not query chirp"})
		return
	}

	//check if the chirp belongs to the user
	if dbChirp.UserID != userId {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "Not authorized to delete this chirp"})
		return
	}

	//delete chirp
	err = cfg.dbQueries.DeleteChirp(r.Context(), validChirpID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Could not delete chirp"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
