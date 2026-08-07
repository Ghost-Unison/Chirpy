package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/alexedwards/argon2id"
)

//哈希是为了存储安全，HTTP负责传输安全

// 使用Argon2id算法对密码进行哈希处理
func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// 将用户输入的密码与数据库中存储的hash密码比较
func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

// 创建JWT
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

// 验证JWT
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	//提供空的claims结构体指针，jwt库解析完之后会把payload中的字段填进去
	claims := &jwt.RegisteredClaims{}

	//返回密钥的回调函数 - 返回的密钥类型必须与签名时一致
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	}

	// 解析JWT，token中包含全部信息
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	// 过期/签名错误都在这返回
	if err != nil {
		return uuid.Nil, err
	}

	// token.Claims 是 Claims 接口类型，需要断言为 *jwt.RegisteredClaims 才能取出其中的字段（与ParseWithClaims中第二个参数一致）
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims type")
	}

	// Subject 中存的是用户 ID 的字符串形式，用 uuid.Parse 转回 uuid.UUID
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

/*

var claims *jwt.RegisteredClaims
—— 只声明了一个指针变量，值是 nil，它不指向任何实际的 RegisteredClaims 结构体内存。就像你有一张地址卡但上面没写地址。
claims := &jwt.RegisteredClaims{}
—— 先 jwt.RegisteredClaims{} 真正分配一个空结构体，再让 claims 指向它。这张地址卡指向一栋真实存在的空房子。

*/

// 从HTTP请求头中获取Bearer Token，用于JWT验证
func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")
	if authorization == "" {
		return "", fmt.Errorf("authorization header is empty")
	}
	if !strings.HasPrefix(authorization, "Bearer ") {
		return "", fmt.Errorf("authorization header does not start with 'Bearer '")
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	return token, nil
}
