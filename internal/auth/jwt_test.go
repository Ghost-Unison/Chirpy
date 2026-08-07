package auth

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// 测试创建 JWT 后能正确验证并取回用户 ID
func TestMakeJWTAndValidate(t *testing.T) {
	tokenSecret := "my-super-secret-key"
	userID := uuid.New()

	token, err := MakeJWT(userID, tokenSecret)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("MakeJWT returned an empty token")
	}

	parsedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT returned unexpected error: %v", err)
	}
	if parsedUserID != userID {
		t.Errorf("expected userID %v, got %v", userID, parsedUserID)
	}
}

// 测试过期的 token 应该被拒绝
func TestValidateJWT_ExpiredToken(t *testing.T) {
	tokenSecret := "my-super-secret-key"
	userID := uuid.New()

	// expiresIn 设为负数，令 ExpiresAt 落在当前时间之前，token 一经签发即过期
	token, err := MakeJWT(userID, tokenSecret)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

// 测试用错误 secret 验证 token 应该被拒绝
func TestValidateJWT_WrongSecret(t *testing.T) {
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	userID := uuid.New()

	token, err := MakeJWT(userID, correctSecret)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

// 测试获取header中的token
func TestGetTokenFromHeader(t *testing.T) {
	token, err := GetBearerToken(http.Header{"Authorization": []string{"Bearer " + "test_token"}})
	if err != nil {
		t.Fatalf("GetBearerToken returned unexpected error: %v", err)
	}
	if token != "test_token" {
		t.Error("expected token 'test_token', got", token)
	}
}
