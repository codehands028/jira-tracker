package models

import (
	"time"
)

// Session 会话表
// 用于存储用户登录后的会话信息，包括Token、过期时间等
type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                                    // 主键ID
	UserID    uint      `gorm:"not null;index:idx_user_id" json:"user_id"`               // 用户ID，关联用户表
	Token     string    `gorm:"type:varchar(500);uniqueIndex;not null" json:"token"`    // JWT令牌，用于身份验证
	SessionID string    `gorm:"type:varchar(100);not null" json:"session_id"`           // 会话唯一标识
	IP        string    `gorm:"type:varchar(50)" json:"ip"`                             // 用户登录IP地址
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`                     // 用户浏览器/客户端信息
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                       // 会话创建时间
	ExpiresAt time.Time `gorm:"index:idx_expires_at" json:"expires_at"`                  // 会话过期时间，用于定时清理过期会话
	IsActive  bool      `gorm:"default:true;index:idx_is_active" json:"is_active"`       // 会话是否激活，true-激活 false-已失效
}
