package services

import (
	"time"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"gorm.io/gorm"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService() *SessionService {
	return &SessionService{
		db: database.DB,
	}
}

// CreateSession 创建会话
func (s *SessionService) CreateSession(userID uint, token, sessionID, ip, userAgent string) error {
	session := &models.Session{
		UserID:    userID,
		Token:     token,
		SessionID:  sessionID,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(config.GlobalConfig.JWT.Expire),
		IsActive:  true,
	}

	return s.db.Create(session).Error
}

// ValidateSession 验证会话是否有效
func (s *SessionService) ValidateSession(token string) (bool, error) {
	var session models.Session
	err := s.db.Where("token = ? AND is_active = ? AND expires_at > ?", token, true, time.Now()).First(&session).Error
	if err != nil {
		return false, err
	}
	return session.IsActive, nil
}

// InvalidateSession 使会话失效
func (s *SessionService) InvalidateSession(token string) error {
	// 获取session信息
	var session models.Session
	if err := s.db.Where("token = ?", token).First(&session).Error; err != nil {
		return err
	}

	// 使CSRF token失效
	if err := s.InvalidateCSRFToken(session.SessionID); err != nil {
		return err
	}

	// 使会话失效
	return s.db.Model(&models.Session{}).Where("token = ?", token).Update("is_active", false).Error
}

// InvalidateAllUserSessions 使用户所有会话失效
func (s *SessionService) InvalidateAllUserSessions(userID uint) error {
	return s.db.Model(&models.Session{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// CleanupExpiredSessions 清理过期会话
func (s *SessionService) CleanupExpiredSessions() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

// GetUserActiveSessions 获取用户活跃会话数
func (s *SessionService) GetUserActiveSessions(userID uint) (int64, error) {
	var count int64
	err := s.db.Model(&models.Session{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&count).Error
	return count, err
}

// InvalidateCSRFToken 使CSRF token失效
func (s *SessionService) InvalidateCSRFToken(sessionID string) error {
	return s.db.Model(&models.Session{}).Where("session_id = ?", sessionID).Update("is_active", false).Error
}
