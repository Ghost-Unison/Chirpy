// Package platform 提供跨业务包共享的 HTTP 基础设施: 统一的 JSON 响应写入与错误响应结构。
package platform

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse 是统一的错误响应结构，格式为 {"error": "..."}
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON marshals payload to JSON and writes it with the given status code.
// 先 marshal 再 WriteHeader，避免重复调用 WriteHeader 导致状态码被忽略。
func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		// marshal 失败几乎不可能发生，兜底返回 500 且不写 body。
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}
