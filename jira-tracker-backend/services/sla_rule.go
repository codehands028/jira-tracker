package services

import (
	"encoding/json"
	"errors"
	"time"

	"jira-tracker-backend/database"
	"jira-tracker-backend/models"
)

type SLARuleService struct{}

func NewSLARuleService() *SLARuleService {
	return &SLARuleService{}
}

// GetSLARules 获取所有SLA规则
func (s *SLARuleService) GetSLARules() ([]models.SLARule, error) {
	var rules []models.SLARule
	if err := database.DB.Order("priority_order desc, created_at desc").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetSLARule 获取单个SLA规则
func (s *SLARuleService) GetSLARule(id uint) (*models.SLARule, error) {
	var rule models.SLARule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return nil, errors.New("规则不存在")
	}
	return &rule, nil
}

// CreateSLARule 创建SLA规则
func (s *SLARuleService) CreateSLARule(req *CreateSLARuleRequest) (*models.SLARule, error) {
	// 计算优先级
	priorityOrder := s.calculatePriorityOrder(req.TicketType, req.Priority, req.Project)

	rule := &models.SLARule{
		Name:          req.Name,
		TicketType:    req.TicketType,
		Priority:      req.Priority,
		Project:       req.Project,
		NormalLimit:   time.Duration(req.NormalLimitHours) * time.Hour,
		SevereLimit:   time.Duration(req.SevereLimitHours) * time.Hour,
		IsActive:      true,
		PriorityOrder: priorityOrder,
	}

	if err := database.DB.Create(rule).Error; err != nil {
		return nil, err
	}

	return rule, nil
}

// UpdateSLARule 更新SLA规则
func (s *SLARuleService) UpdateSLARule(id uint, req *UpdateSLARuleRequest) (*models.SLARule, error) {
	var rule models.SLARule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return nil, errors.New("规则不存在")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.TicketType != nil {
		updates["ticket_type"] = *req.TicketType
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Project != nil {
		updates["project"] = *req.Project
	}
	if req.NormalLimitHours > 0 {
		updates["normal_limit"] = time.Duration(req.NormalLimitHours) * time.Hour
	}
	if req.SevereLimitHours > 0 {
		updates["severe_limit"] = time.Duration(req.SevereLimitHours) * time.Hour
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// 如果关键字段变化，重新计算优先级
	if req.TicketType != nil || req.Priority != nil || req.Project != nil {
		ticketType := rule.TicketType
		priority := rule.Priority
		project := rule.Project
		if req.TicketType != nil {
			ticketType = *req.TicketType
		}
		if req.Priority != nil {
			priority = *req.Priority
		}
		if req.Project != nil {
			project = *req.Project
		}
		updates["priority_order"] = s.calculatePriorityOrder(ticketType, priority, project)
	}

	if err := database.DB.Model(&rule).Updates(updates).Error; err != nil {
		return nil, err
	}

	database.DB.First(&rule, id)
	return &rule, nil
}

// DeleteSLARule 删除SLA规则
func (s *SLARuleService) DeleteSLARule(id uint) error {
	var rule models.SLARule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return errors.New("规则不存在")
	}

	return database.DB.Delete(&rule).Error
}

// ToggleSLARule 启用/禁用SLA规则
func (s *SLARuleService) ToggleSLARule(id uint, isActive bool) error {
	var rule models.SLARule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return errors.New("规则不存在")
	}

	return database.DB.Model(&rule).Update("is_active", isActive).Error
}

// MatchSLARule 根据工单属性匹配最合适的SLA规则
func (s *SLARuleService) MatchSLARule(priority string) (*models.SLARule, error) {
	var rule models.SLARule

	// 优先匹配：优先级 + 项目 + 类型
	// 然后匹配：优先级 + 项目
	// 然后匹配：优先级 + 类型
	// 然后匹配：仅优先级
	// 最后：默认规则
	query := database.DB.Where("is_active = ?", true)

	// 按优先级匹配
	err := query.Where("priority = ? OR priority = ''", priority).
		Order("priority_order desc, created_at desc").
		First(&rule).Error

	if err != nil {
		// 如果没有匹配的规则，尝试获取默认规则
		err = database.DB.Where("is_active = ? AND priority = ''", true).
			Order("priority_order desc, created_at desc").
			First(&rule).Error
		if err != nil {
			return nil, errors.New("未找到匹配的SLA规则")
		}
	}

	return &rule, nil
}

// calculatePriorityOrder 计算规则匹配优先级
// 规则越具体，优先级越高
func (s *SLARuleService) calculatePriorityOrder(ticketType, priority, project string) int {
	order := 0
	if ticketType != "" {
		order += 100
	}
	if priority != "" {
		order += 10
	}
	if project != "" {
		order += 1
	}
	return order
}

// GetSLAStatus 获取工单的SLA状态
func (s *SLARuleService) GetSLAStatus(ticketID uint, priority string, lastFlowTime time.Time) (*SLAStatus, error) {
	rule, err := s.MatchSLARule(priority)
	if err != nil {
		return nil, err
	}

	stayDuration := time.Since(lastFlowTime)
	remainingTime := rule.NormalLimit - stayDuration

	status := &SLAStatus{
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		NormalLimit:   int64(rule.NormalLimit.Seconds()),
		SevereLimit:   int64(rule.SevereLimit.Seconds()),
		StayDuration:  int64(stayDuration.Seconds()),
		RemainingTime: int64(remainingTime.Seconds()),
		IsTimeout:     stayDuration > rule.NormalLimit,
		TimeoutLevel:  "",
		Progress:      float64(stayDuration) / float64(rule.NormalLimit) * 100,
	}

	if stayDuration > rule.SevereLimit {
		status.TimeoutLevel = "severe"
		status.Progress = 100
	} else if stayDuration > rule.NormalLimit {
		status.TimeoutLevel = "normal"
		status.Progress = 100
	}

	return status, nil
}

// CreateSLARuleRequest 创建SLA规则请求
type CreateSLARuleRequest struct {
	Name             string `json:"name" binding:"required"`
	TicketType       string `json:"ticket_type"`
	Priority         string `json:"priority"`
	Project          string `json:"project"`
	NormalLimitHours int    `json:"normal_limit_hours" binding:"required,min=1"`
	SevereLimitHours int    `json:"severe_limit_hours" binding:"required,min=1"`
}

// UpdateSLARuleRequest 更新SLA规则请求
type UpdateSLARuleRequest struct {
	Name             string  `json:"name"`
	TicketType       *string `json:"ticket_type"`
	Priority         *string `json:"priority"`
	Project          *string `json:"project"`
	NormalLimitHours int     `json:"normal_limit_hours"`
	SevereLimitHours int     `json:"severe_limit_hours"`
	IsActive         *bool   `json:"is_active"`
}

// SLAStatus SLA状态
type SLAStatus struct {
	RuleID        uint    `json:"rule_id"`
	RuleName      string  `json:"rule_name"`
	NormalLimit   int64   `json:"normal_limit"`    // 普通时限（秒）
	SevereLimit   int64   `json:"severe_limit"`    // 严重时限（秒）
	StayDuration  int64   `json:"stay_duration"`   // 已停留时长（秒）
	RemainingTime int64   `json:"remaining_time"`  // 剩余时间（秒），负数表示已超时
	IsTimeout     bool    `json:"is_timeout"`
	TimeoutLevel  string  `json:"timeout_level"`
	Progress      float64 `json:"progress"` // SLA进度百分比
}

// BatchOperationDetail 批量操作详情
type BatchOperationDetail struct {
	ToUserID   uint   `json:"to_user_id,omitempty"`
	Content    string `json:"content,omitempty"`
	Conclusion string `json:"conclusion,omitempty"`
}

// CreateBatchOperationLog 创建批量操作日志
func CreateBatchOperationLog(userID uint, operationType string, ticketIDs []uint, successCount int, detail *BatchOperationDetail, ipAddress string) error {
	ticketIDsJSON, _ := json.Marshal(ticketIDs)
	detailJSON, _ := json.Marshal(detail)

	log := &models.BatchOperationLog{
		UserID:        userID,
		OperationType: operationType,
		TicketIDs:     string(ticketIDsJSON),
		TicketCount:   len(ticketIDs),
		SuccessCount:  successCount,
		Detail:        string(detailJSON),
		IPAddress:     ipAddress,
	}

	return database.DB.Create(log).Error
}

// GetBatchOperationLogs 获取批量操作日志
func GetBatchOperationLogs(page, pageSize int, userID uint, operationType string) ([]models.BatchOperationLog, int64, error) {
	var logs []models.BatchOperationLog
	var total int64

	query := database.DB.Model(&models.BatchOperationLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if operationType != "" {
		query = query.Where("operation_type = ?", operationType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
