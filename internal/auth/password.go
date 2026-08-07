package auth

import (
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
