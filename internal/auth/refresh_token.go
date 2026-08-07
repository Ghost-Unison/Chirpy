package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// generate a random 256-bis(32-byte) hex-encoded string
func MakeRefreshToken() string {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	return hex.EncodeToString(randBytes)
}
