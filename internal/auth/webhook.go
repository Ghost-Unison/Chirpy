package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// 从http请求头中获取API密钥，用于webhook验证
func GetAPIKey(headers http.Header) (string, error) {
	rawApiKey := headers.Get("Authorization")
	if rawApiKey == "" {
		return "", fmt.Errorf("authorization header is empty")
	}
	if !strings.HasPrefix(rawApiKey, "ApiKey ") {
		return "", fmt.Errorf("authorization header does not start with 'ApiKey '")
	}
	token := strings.TrimPrefix(rawApiKey, "ApiKey ")
	return token, nil
}
