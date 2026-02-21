package services

import (
	"context"
	"fmt"
	"math/rand"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"

	"github.com/redis/go-redis/v9"
)

type SMSService struct{}

func NewSMSService() *SMSService {
	return &SMSService{}
}

// SendCode 发送验证码
func (s *SMSService) SendCode(phone string) error {
	ctx := context.Background()

	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(900000)+100000)

	// 存储到Redis，设置过期时间
	key := fmt.Sprintf("sms:code:%s", phone)
	err := database.RedisClient.Set(ctx, key, code, config.GlobalConfig.SMS.Expire).Err()
	if err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	// 实际项目中这里调用短信服务商API发送短信
	// 示例中仅打印日志
	fmt.Printf("SMS sent to %s with code: %s\n", phone, code)

	return nil
}

// VerifyCode 验证验证码
func (s *SMSService) VerifyCode(phone, code string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("sms:code:%s", phone)

	storedCode, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("failed to get verification code: %w", err)
	}

	// 验证成功后删除验证码
	if storedCode == code {
		database.RedisClient.Del(ctx, key)
		return true, nil
	}

	return false, nil
}
