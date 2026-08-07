package chirps

import (
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/Ghost-Unison/Chirpy/internal/auth"
	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/Ghost-Unison/Chirpy/internal/platform"
	"github.com/google/uuid"
)

// Handler 持有 chirp 相关处理器所需的依赖。
type Handler struct {
	db        *database.Queries
	secretKey string
}

// NewHandler 构造一个持有 db 与 JWT 密钥的 chirps.Handler。
func NewHandler(db *database.Queries, secretKey string) *Handler {
	return &Handler{db: db, secretKey: secretKey}
}

// Create 处理 POST /api/chirps: 鉴权后创建一条 chirp。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// jwt auth
	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{
			Error: "Invalid token",
		})
		return
	}
	userId, err := auth.ValidateJWT(jwt, h.secretKey)
	if err != nil {
		platform.WriteJSON(w, http.StatusUnauthorized, platform.ErrorResponse{
			Error: "Invalid token",
		})
		return
	}

	// decode request
	var req chirpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{
			Error: "Invalid request payload",
		})
		return
	}

	// length judge（按 rune 计数，即字符数而非字节数）
	if utf8.RuneCountInString(req.Body) > 140 {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{
			Error: "Body length cannot exceed 140 characters",
		})
		return
	}
	/*
		validate user ID -当 json.NewDecoder().Decode() 解析 JSON 时，会自动尝试把 user_id 字段的字符串解析成 UUID，这里只需要排除空值
		if req.UserId == uuid.Nil {
			platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{
				Error: "User ID is required",
			})
			return
		}
	*/

	//不再返回 cleaned_body，而是存入数据库
	cleanedBody := cleanBody(req.Body)
	dbChirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	})
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{
			Error: "Could not create chirp",
		})
		return
	}

	// 成功：返回 201 + 完整 chirp 资源
	platform.WriteJSON(w, http.StatusCreated, databaseChirp(dbChirp))
}

// List 处理 GET /api/chirps: 支持按 author_id 过滤、sort 排序。
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// author_id 是可选查询参数：未提供时返回全部 chirps，提供时只返回该作者的 chirps。 - 过滤在数据库层面完成，而非在内存中。
	rawAuthorID := r.URL.Query().Get("author_id")
	// sort 是可选查询参数：asc（默认）/desc，按 created_at 排序
	sortParam := r.URL.Query().Get("sort")

	var dbChirps []database.Chirp
	var err error
	if rawAuthorID == "" {
		// 未提供 author_id：返回所有 chirps，按 created_at 升序
		dbChirps, err = h.db.QueryChirps(r.Context())
	} else {
		// 提供了 author_id：解析后按作者过滤（数据库层面）
		authorID, parseErr := uuid.Parse(rawAuthorID)
		if parseErr != nil {
			platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Invalid author ID"})
			return
		}
		dbChirps, err = h.db.QueryChirpsByAuthor(r.Context(), authorID)
	}
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Could not query chirps"})
		return
	}

	// 在内存中按 created_at 排序：默认 asc，sort=desc 时倒序
	sortChirps(dbChirps, sortParam, "created_at")

	platform.WriteJSON(w, http.StatusOK, databaseChirps(dbChirps))
}

// Get 处理 GET /api/chirps/{chirpID}: 按主键查询单条 chirp。
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	chirpID := r.PathValue("chirpID")
	validChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{
			Error: "Invalid chirp ID",
		})
		return
	}

	dbChirp, err := h.db.QueryChirp(r.Context(), validChirpID)
	if err != nil {
		platform.WriteJSON(w, http.StatusNotFound, platform.ErrorResponse{
			Error: "Could not query chirp",
		})
		return
	}

	platform.WriteJSON(w, http.StatusOK, databaseChirp(dbChirp))
}

// Delete 处理 DELETE /api/chirps/{chirpID}: 鉴权后仅允许作者删除自己的 chirp。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	//get chirp
	chirpID := r.PathValue("chirpID")
	validChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		platform.WriteJSON(w, http.StatusBadRequest, platform.ErrorResponse{Error: "Invalid chirp ID"})
		return
	}

	dbChirp, err := h.db.QueryChirp(r.Context(), validChirpID)
	if err != nil {
		platform.WriteJSON(w, http.StatusNotFound, platform.ErrorResponse{Error: "Could not query chirp"})
		return
	}

	//check if the chirp belongs to the user
	if dbChirp.UserID != userId {
		platform.WriteJSON(w, http.StatusForbidden, platform.ErrorResponse{Error: "Not authorized to delete this chirp"})
		return
	}

	//delete chirp
	err = h.db.DeleteChirp(r.Context(), validChirpID)
	if err != nil {
		platform.WriteJSON(w, http.StatusInternalServerError, platform.ErrorResponse{Error: "Could not delete chirp"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
