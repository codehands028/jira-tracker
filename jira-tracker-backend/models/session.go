package models

import (
	"time"
)

// Session 会话表
type Session struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index:idx_user_id"`
	Token     string    `gorm:"type:varchar(500);uniqueIndex;not null"`
	SessionID string    `gorm:"type:varchar(100);not null"`
	IP        string    `gorm:"type:varchar(50)"`
	UserAgent string    `gorm:"type:varchar(500)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	ExpiresAt time.Time `gorm:"index:idx_expires_at"`
	IsActive  bool      `gorm:"default:true;index:idx_is_active"`
}
