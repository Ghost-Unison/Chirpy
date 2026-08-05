package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"
)

type validateResponse struct {
	Valid       bool   `json:"valid"`
	Error       string `json:"error"`
	CleanedBody string `json:"cleaned_body"`
}

type validateRequest struct {
	Body string `json:"body"`
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

/*
//用数组来判断的写法
var badWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanBody(body string) string {
    words := strings.Split(body, " ")
    for i, word := range words {
        lowered := strings.ToLower(word)
		//*遍历方法1
        for _, bw := range badWords {
            if lowered == bw {
                words[i] = "****"
                break // 一个单词最多命中一个敏感词,命中即跳出,避免无意义的后续比较
            }
        }
		//*遍历方法2
		if slices.Contains(badWords, strings.ToLower(word)) {
            words[i] = "****"
        }
    }
    return strings.Join(words, " ")
}
*/

// writeJSON marshals payload to JSON and writes it with the given status code.
// 先 marshal 再 WriteHeader，避免重复调用 WriteHeader 导致状态码被忽略。
func writeJSON(w http.ResponseWriter, statusCode int, payload validateResponse) {
	data, err := json.Marshal(payload)
	if err != nil {
		// marshal 失败几乎不可能发生（结构体只含 bool/string），
		// 兜底返回 500 且不写 body，避免再 marshal 触发同样的错误。
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}

/*
接收一个json格式的参数，判断其中body属性的长度不能超过140个字符

如果发生错误，则返回合适的statusCode和json格式的错误信息

	{
	  "error": "{error message}"
	}

如果成功，则返回statusCode 200和json格式的正确信息

	{
	  "valid": true,
	  "cleaned_body": "{cleaned text}"
	}

敏感词校验：body 中的敏感词（独立单词、大小写不敏感）会被替换为 ****，
cleaned_body 始终返回（即使没有任何敏感词，cleaned_body 即原文）。
*/
func validateFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// decode request
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, validateResponse{
			Error: "Invalid request payload: " + err.Error(),
		})
		return
	}

	// length judge（按 rune 计数，即字符数而非字节数）
	if utf8.RuneCountInString(req.Body) > 140 {
		writeJSON(w, http.StatusBadRequest, validateResponse{
			Error: "Body length cannot exceed 140 characters",
		})
		return
	}

	// 总是返回 cleaned 版本（无敏感词时 cleaned_body 即原文）
	writeJSON(w, http.StatusOK, validateResponse{
		Valid:       true,
		CleanedBody: cleanBody(req.Body),
	})
}
