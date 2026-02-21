package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Phone     string `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	Name      string `gorm:"type:varchar(50);not null" json:"name"`
	Role      string `gorm:"type:varchar(20);not null;comment:admin/test/dev" json:"role"` // 管理员/测试/研发
	Status    int    `gorm:"type:tinyint;default:1;comment:1-启用 0-禁用" json:"status"`
	TicketNum int    `gorm:"type:int;default:0;comment:当前待处理工单数" json:"ticket_num"`
}

// Ticket 工单表
type Ticket struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	JiraKey       string `gorm:"type:varchar(50);uniqueIndex;not null" json:"jira_key"`
	JiraURL       string `gorm:"type:varchar(500);not null" json:"jira_url"`
	Description   string `gorm:"type:text" json:"description"`
	Priority      string `gorm:"type:varchar(20);comment:low/medium/high/critical" json:"priority"`
	Status        string `gorm:"type:varchar(20);not null;comment:processing/retesting/closed/timeout" json:"status"`
	CurrentUserID uint   `gorm:"not null;index" json:"current_user_id"`
	CurrentUser   User   `gorm:"foreignKey:CurrentUserID" json:"current_user"`
	IsTimeout     bool   `gorm:"type:tinyint;default:0" json:"is_timeout"`
	TimeoutLevel  string `gorm:"type:varchar(20);comment:normal/severe" json:"timeout_level"`
}

// TicketFlow 工单流转记录表
type TicketFlow struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	TicketID     uint         `gorm:"not null;index" json:"ticket_id"`
	Ticket       Ticket       `gorm:"foreignKey:TicketID" json:"ticket"`
	FromUserID   uint         `gorm:"not null" json:"from_user_id"`
	FromUser     User         `gorm:"foreignKey:FromUserID" json:"from_user"`
	ToUserID     uint         `gorm:"not null" json:"to_user_id"`
	ToUser       User         `gorm:"foreignKey:ToUserID" json:"to_user"`
	Content      string       `gorm:"type:text" json:"content"`
	ProcessTime  time.Duration `gorm:"type:bigint;comment:处理时长(秒)" json:"process_time"`
	IsTimeout    bool         `gorm:"type:tinyint;default:0" json:"is_timeout"`
	TimeoutLevel string       `gorm:"type:varchar(20);comment:normal/severe" json:"timeout_level"`
}

// Notification 通知表
type Notification struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID   uint   `gorm:"not null;index" json:"user_id"`
	User     User   `gorm:"foreignKey:UserID" json:"user"`
	Type     string `gorm:"type:varchar(50);not null;comment:timeout/flow/close" json:"type"`
	Title    string `gorm:"type:varchar(200);not null" json:"title"`
	Content  string `gorm:"type:text" json:"content"`
	IsRead   bool   `gorm:"type:tinyint;default:0" json:"is_read"`
	TicketID uint   `gorm:"index" json:"ticket_id"`
	Ticket   Ticket `gorm:"foreignKey:TicketID" json:"ticket"`
}

// OperationLog 操作日志表
type OperationLog struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	User        User   `gorm:"foreignKey:UserID" json:"user"`
	Module      string `gorm:"type:varchar(50);not null" json:"module"`
	Action      string `gorm:"type:varchar(50);not null" json:"action"`
	Description string `gorm:"type:text" json:"description"`
	IPAddress   string `gorm:"type:varchar(50)" json:"ip_address"`
}

// TimeoutRule 超时规则表
type TimeoutRule struct {
	ID        uint          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string        `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	NormalLimit time.Duration `gorm:"type:bigint;not null;comment:普通超时时间(秒)" json:"normal_limit"`
	SevereLimit time.Duration `gorm:"type:bigint;not null;comment:严重超时时间(秒)" json:"severe_limit"`
	IsActive    bool          `gorm:"type:tinyint;default:0" json:"is_active"`
}
