package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"jira-tracker-backend/database"
)

var ctx = context.Background()

const (
	// CSRFTokenLength CSRF token长度
	CSRFTokenLength = 32
	// CSRFTokenPrefix Redis中CSRF token的key前缀
	CSRFTokenPrefix = "csrf:"
	// CSRFTokenExpiration CSRF token过期时间（24小时）
	CSRFTokenExpiration = 24 * time.Hour
)

var (
	// ErrInvalidCSRFToken CSRF token无效错误
	ErrInvalidCSRFToken = errors.New("无效的CSRF token")
	// ErrCSRFTokenExpired CSRF token已过期错误
	ErrCSRFTokenExpired = errors.New("CSRF token已过期")
)

// GenerateCSRFToken 生成CSRF Token
func GenerateCSRFToken(sessionID string) (string, error) {
	// 生成随机字节
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成CSRF token失败: %w", err)
	}

	// 将字节转换为十六进制字符串
	token := hex.EncodeToString(bytes)

	// 将token存储到Redis中，key格式为: csrf:sessionID
	key := fmt.Sprintf("%s%s", CSRFTokenPrefix, sessionID)
	if err := database.RedisClient.Set(ctx, key, token, CSRFTokenExpiration).Err(); err != nil {
		return "", fmt.Errorf("存储CSRF token失败: %w", err)
	}

	return token, nil
}

// ValidateCSRFToken 验证CSRF Token
func ValidateCSRFToken(sessionID, token string) error {
	if sessionID == "" || token == "" {
		return ErrInvalidCSRFToken
	}

	// 从Redis中获取存储的token
	key := fmt.Sprintf("%s%s", CSRFTokenPrefix, sessionID)
	storedToken, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return ErrCSRFTokenExpired
		}
		return fmt.Errorf("获取CSRF token失败: %w", err)
	}

	// 比对token
	if storedToken != token {
		return ErrInvalidCSRFToken
	}

	return nil
}

// RefreshCSRFToken 刷新CSRF Token
func RefreshCSRFToken(sessionID string) (string, error) {
	// 先删除旧的token
	key := fmt.Sprintf("%s%s", CSRFTokenPrefix, sessionID)
	database.RedisClient.Del(ctx, key)

	// 生成新的token
	return GenerateCSRFToken(sessionID)
}

// InvalidateCSRFToken 使CSRF Token失效
func InvalidateCSRFToken(sessionID string) error {
	key := fmt.Sprintf("%s%s", CSRFTokenPrefix, sessionID)
	return database.RedisClient.Del(ctx, key).Err()
}
