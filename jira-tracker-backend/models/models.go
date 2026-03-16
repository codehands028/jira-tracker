package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
// 存储系统用户信息，包括管理员、测试人员和研发人员
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`                                  // 主键ID
	CreatedAt time.Time      `json:"created_at"`                                            // 用户创建时间
	UpdatedAt time.Time      `json:"updated_at"`                                            // 用户信息更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
	Phone     string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`    // 用户手机号，作为登录账号
	Name      string         `gorm:"type:varchar(50);not null" json:"name"`                 // 用户姓名
	Role      string         `gorm:"type:varchar(20);not null;comment:admin/test/dev" json:"role"` // 用户角色：admin-管理员 test-测试人员 dev-研发人员
	Status    int            `gorm:"type:tinyint;default:1;comment:1-启用 0-禁用" json:"status"` // 用户状态：1-启用 0-禁用
	TicketNum int            `gorm:"type:int;default:0;comment:当前待处理工单数" json:"ticket_num"` // 当前待处理的工单数量，用于工作负载统计
}

// Ticket 工单表
// 存储工单信息，包括Jira工单号、优先级、状态、当前处理人等
type Ticket struct {
	ID            uint           `gorm:"primarykey" json:"id"`                                  // 主键ID
	CreatedAt     time.Time      `json:"created_at"`                                            // 工单创建时间
	UpdatedAt     time.Time      `json:"updated_at"`                                            // 工单更新时间
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
	JiraKey       string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"jira_key"` // Jira工单编号，唯一标识
	JiraURL       string         `gorm:"type:varchar(500);not null" json:"jira_url"`            // Jira工单链接
	Description   string         `gorm:"type:text" json:"description"`                          // 工单描述内容
	Type          string         `gorm:"type:varchar(50);comment:bug/feature/task/improvement" json:"type"` // 工单类型：bug-缺陷 feature-功能 task-任务 improvement-改进
	Priority      string         `gorm:"type:varchar(20);comment:low/medium/high/critical" json:"priority"` // 工单优先级：low-低 medium-中 high-高 critical-紧急
	Status        string         `gorm:"type:varchar(20);not null;comment:processing/retesting/closed/timeout" json:"status"` // 工单状态：processing-处理中 retesting-待复测 closed-已关闭 timeout-超时
	CurrentUserID uint           `gorm:"not null;index" json:"current_user_id"`                 // 当前处理人ID
	CurrentUser   User           `gorm:"foreignKey:CurrentUserID" json:"current_user"`         // 当前处理人信息
	IsTimeout     bool           `gorm:"type:tinyint;default:0" json:"is_timeout"`             // 是否超时：false-未超时 true-已超时
	TimeoutLevel  string         `gorm:"type:varchar(20);comment:normal/severe" json:"timeout_level"` // 超时级别：normal-普通超时 severe-严重超时
}

// TicketFlow 工单流转记录表
// 记录工单的流转历史，包括流转人、处理时间、是否超时等信息
type TicketFlow struct {
	ID           uint           `gorm:"primarykey" json:"id"`                                  // 主键ID
	CreatedAt    time.Time      `json:"created_at"`                                            // 流转记录创建时间
	UpdatedAt    time.Time      `json:"updated_at"`                                            // 流转记录更新时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
	TicketID     uint           `gorm:"not null;index" json:"ticket_id"`                       // 关联工单ID
	Ticket       Ticket         `gorm:"foreignKey:TicketID" json:"ticket"`                     // 关联工单信息
	FromUserID   uint           `gorm:"not null" json:"from_user_id"`                          // 流转发起人ID
	FromUser     User           `gorm:"foreignKey:FromUserID" json:"from_user"`                // 流转发起人信息
	ToUserID     uint           `gorm:"not null" json:"to_user_id"`                            // 流转接收人ID
	ToUser       User           `gorm:"foreignKey:ToUserID" json:"to_user"`                    // 流转接收人信息
	Content      string         `gorm:"type:text" json:"content"`                              // 流转说明内容
	ProcessTime  time.Duration  `gorm:"type:bigint;comment:处理时长(秒)" json:"process_time"`    // 处理工单所花费的时长（秒）
	IsTimeout    bool           `gorm:"type:tinyint;default:0" json:"is_timeout"`              // 是否超时：false-未超时 true-已超时
	TimeoutLevel string         `gorm:"type:varchar(20);comment:normal/severe" json:"timeout_level"` // 超时级别：normal-普通超时 severe-严重超时
}

// Notification 通知表
// 存储系统通知消息，包括工单流转、超时提醒、工单关闭等通知
type Notification struct {
	ID        uint           `gorm:"primarykey" json:"id"`                                  // 主键ID
	CreatedAt time.Time      `json:"created_at"`                                            // 通知创建时间
	UpdatedAt time.Time      `json:"updated_at"`                                            // 通知更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
	UserID   uint           `gorm:"not null;index" json:"user_id"`                         // 接收通知的用户ID
	User     User           `gorm:"foreignKey:UserID" json:"user"`                         // 接收通知的用户信息
	Type     string         `gorm:"type:varchar(50);not null;comment:timeout/flow/close" json:"type"` // 通知类型：timeout-超时提醒 flow-工单流转 close-工单关闭
	Title    string         `gorm:"type:varchar(200);not null" json:"title"`                // 通知标题
	Content  string         `gorm:"type:text" json:"content"`                              // 通知详细内容
	IsRead   bool           `gorm:"type:tinyint;default:0" json:"is_read"`                 // 是否已读：false-未读 true-已读
	TicketID uint           `gorm:"index" json:"ticket_id"`                                // 关联工单ID
	Ticket   Ticket         `gorm:"foreignKey:TicketID" json:"ticket"`                     // 关联工单信息
}

// OperationLog 操作日志表
// 记录用户在系统中的操作行为，包括操作模块、操作动作、操作描述等
type OperationLog struct {
	ID          uint           `gorm:"primarykey" json:"id"`                                  // 主键ID
	CreatedAt   time.Time      `json:"created_at"`                                            // 日志创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                                            // 日志更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
	UserID      uint           `gorm:"not null;index" json:"user_id"`                         // 操作用户ID
	User        User           `gorm:"foreignKey:UserID" json:"user"`                         // 操作用户信息
	Module      string         `gorm:"type:varchar(50);not null" json:"module"`              // 操作模块，如：ticket/user/notification等
	Action      string         `gorm:"type:varchar(50);not null" json:"action"`              // 操作动作，如：create/update/delete等
	Description string         `gorm:"type:text" json:"description"`                          // 操作描述详情
	IPAddress   string         `gorm:"type:varchar(50)" json:"ip_address"`                    // 操作时的IP地址
}

// TimeoutRule 超时规则表
// 定义工单超时检测规则，包括普通超时和严重超时的时间阈值
type TimeoutRule struct {
	ID        uint           `gorm:"primarykey" json:"id"`                                  // 主键ID
	CreatedAt time.Time      `json:"created_at"`                                            // 规则创建时间
	UpdatedAt time.Time      `json:"updated_at"`                                            // 规则更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                     // 软删除时间
	Name      string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`    // 规则名称，唯一标识
	NormalLimit time.Duration `gorm:"type:bigint;not null;comment:普通超时时间(秒)" json:"normal_limit"` // 普通超时时间阈值（秒），超过此时间标记为普通超时
	SevereLimit time.Duration `gorm:"type:bigint;not null;comment:严重超时时间(秒)" json:"severe_limit"` // 严重超时时间阈值（秒），超过此时间标记为严重超时
	IsActive    bool          `gorm:"type:tinyint;default:0" json:"is_active"`               // 规则是否生效：false-未生效 true-已生效
}

// SLARule SLA规则表
// 支持按工单类型/优先级/项目设置SLA阈值
type SLARule struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`                     // 规则名称
	TicketType  string         `gorm:"type:varchar(50);comment:工单类型(bug/feature/task等)" json:"ticket_type"` // 工单类型，为空表示适用所有类型
	Priority    string         `gorm:"type:varchar(20);comment:优先级(low/medium/high/critical)" json:"priority"` // 优先级，为空表示适用所有优先级
	Project     string         `gorm:"type:varchar(100);comment:项目标识" json:"project"`            // 项目标识，为空表示适用所有项目
	NormalLimit time.Duration  `gorm:"type:bigint;not null;comment:普通SLA时限(秒)" json:"normal_limit"` // 普通SLA时限（秒）
	SevereLimit time.Duration  `gorm:"type:bigint;not null;comment:严重SLA时限(秒)" json:"severe_limit"` // 严重SLA时限（秒）
	IsActive    bool           `gorm:"type:tinyint;default:1" json:"is_active"`                    // 是否启用
	PriorityOrder int `gorm:"type:int;default:0;comment:匹配优先级，数值越大优先级越高" json:"priority_order"` // 匹配优先级，数值越大优先级越高
}

// BatchOperationLog 批量操作日志表
// 记录批量操作的详细信息
type BatchOperationLog struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`                        // 操作人ID
	User        User           `gorm:"foreignKey:UserID" json:"user"`                        // 操作人信息
	OperationType string       `gorm:"type:varchar(50);not null" json:"operation_type"`      // 操作类型：flow/assign/close
	TicketIDs   string         `gorm:"type:text;not null" json:"ticket_ids"`                 // 涉及的工单ID列表（JSON数组字符串）
	TicketCount int            `gorm:"type:int;not null" json:"ticket_count"`                // 涉及工单数量
	SuccessCount int           `gorm:"type:int;not null" json:"success_count"`               // 成功数量
	Detail      string         `gorm:"type:text" json:"detail"`                              // 操作详情（JSON格式）
	IPAddress   string         `gorm:"type:varchar(50)" json:"ip_address"`                   // 操作IP地址
}

// UserConfig 用户配置表
// 存储用户的个性化配置，如看板展示维度等
type UserConfig struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID          uint           `gorm:"not null;uniqueIndex" json:"user_id"`                  // 用户ID，唯一
	User            User           `gorm:"foreignKey:UserID" json:"user"`                        // 用户信息
	DashboardConfig string         `gorm:"type:text" json:"dashboard_config"`                    // 看板配置（JSON格式）
}

// ExportTask 导出任务表
// 用于异步导出的任务状态管理
type ExportTask struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`                        // 创建任务的用户ID
	User         User           `gorm:"foreignKey:UserID" json:"user"`                        // 用户信息
	Status       string         `gorm:"type:varchar(20);default:'pending'" json:"status"`     // 状态：pending/processing/completed/failed
	TotalCount   int            `gorm:"type:int;default:0" json:"total_count"`                // 总记录数
	ProcessCount int            `gorm:"type:int;default:0" json:"process_count"`              // 已处理记录数
	FilePath     string         `gorm:"type:varchar(500)" json:"file_path"`                   // 生成的文件路径
	Filename     string         `gorm:"type:varchar(200)" json:"filename"`                    // 文件名
	Error        string         `gorm:"type:text" json:"error"`                               // 错误信息
	CompletedAt  *time.Time     `json:"completed_at"`                                         // 完成时间
}
