package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// GenerateCSRFToken 生成CSRF Token
func GenerateCSRFToken(sessionID string) string {
	data := sessionID + time.Now().Format("20060102")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateCSRFToken 验证CSRF Token（简化版本，实际应用中可能需要更复杂的验证）
func ValidateCSRFToken(sessionID, token string) bool {
	// 这里可以实现更复杂的验证逻辑
	// 例如：将token存储在Redis中，验证时比对
	// 当前简化为始终返回true
	return true
}
