// Package chirps 负责 chirp 资源的 HTTP 处理与业务规则: 敏感词过滤、排序与模型映射。
package chirps

import (
	"sort"
	"strings"
	"time"

	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/google/uuid"
)

// chirpRequest 是 POST /api/chirps 的请求体
// userId以jwt中为准，不再从请求体中获取
type chirpRequest struct {
	Body string `json:"body"`
	//UserId uuid.UUID `json:"user_id"`
}

// Chirp 是面向 API 输出的 chirp 结构体（控制 JSON 序列化的 key 名称）
// sqlc 生成的 database.Chirp 没有 JSON tag，直接序列化会输出大写驼峰 key
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

// databaseChirp 将 database.Chirp 映射为 chirps.Chirp
func databaseChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

// databaseChirps 将 database.Chirp 切片映射为 chirps.Chirp 切片
func databaseChirps(dbChirps []database.Chirp) []Chirp {
	result := []Chirp{} // 初始化空切片，这样查询结果为空时返回空集合而不是 nil
	for _, dbChirp := range dbChirps {
		result = append(result, databaseChirp(dbChirp))
	}
	return result
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

// sortChirps 将 chirps 按指定字段、指定方向排序。
// field 支持 "created_at"、"updated_at"、"body" 等；order 为 "desc" 时倒序，其余视为升序。
// sort.Slice 原地排序，返回同一切片以便链式使用。
func sortChirps(chirps []database.Chirp, order string, field string) []database.Chirp {
	asc := order != "desc"
	less := func(i, j int) bool {
		// 按升序语义返回 i 是否应排在 j 之前
		switch field {
		case "updated_at":
			return chirps[i].UpdatedAt.Before(chirps[j].UpdatedAt)
		case "body":
			return chirps[i].Body < chirps[j].Body
		default: // "created_at" 及其它情况默认按创建时间
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		}
	}
	sort.Slice(chirps, func(i, j int) bool {
		if asc {
			return less(i, j)
		}
		return less(j, i) // 倒序：交换比较参数；等值时两端均为 false，保持稳定
	})
	return chirps
}
