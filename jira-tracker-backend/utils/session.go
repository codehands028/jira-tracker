package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
