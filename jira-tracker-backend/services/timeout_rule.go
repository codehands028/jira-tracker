package services

import (
	"errors"
	"time"

	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"gorm.io/gorm"
)

type TimeoutRuleService struct{}

func NewTimeoutRuleService() *TimeoutRuleService {
	return &TimeoutRuleService{}
}

// GetTimeoutRules 获取所有超时规则
func (s *TimeoutRuleService) GetTimeoutRules() ([]models.TimeoutRule, error) {
	var rules []models.TimeoutRule
	if err := database.DB.Order("created_at desc").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetActiveTimeoutRule 获取当前生效的超时规则
func (s *TimeoutRuleService) GetActiveTimeoutRule() (*models.TimeoutRule, error) {
	var rule models.TimeoutRule
	if err := database.DB.Where("is_active = ?", true).First(&rule).Error; err != nil {
		return nil, errors.New("未找到有效的超时规则")
	}
	return &rule, nil
}

// CreateTimeoutRule 创建超时规则
func (s *TimeoutRuleService) CreateTimeoutRule(req *CreateTimeoutRuleRequest) (*models.TimeoutRule, error) {
	// 检查名称是否重复
	var count int64
	database.DB.Model(&models.TimeoutRule{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, errors.New("规则名称已存在")
	}

	rule := &models.TimeoutRule{
		Name:        req.Name,
		NormalLimit: time.Duration(req.NormalLimitHours) * time.Hour,
		SevereLimit: time.Duration(req.SevereLimitHours) * time.Hour,
		IsActive:    false, // 新创建的规则默认不生效
	}

	if err := database.DB.Create(rule).Error; err != nil {
		return nil, err
	}

	return rule, nil
}

// UpdateTimeoutRule 更新超时规则
func (s *TimeoutRuleService) UpdateTimeoutRule(id uint, req *UpdateTimeoutRuleRequest) (*models.TimeoutRule, error) {
	var rule models.TimeoutRule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return nil, errors.New("规则不存在")
	}

	updates := map[string]any{}
	if req.Name != "" {
		// 检查名称是否重复
		var count int64
		database.DB.Model(&models.TimeoutRule{}).Where("name = ? AND id != ?", req.Name, id).Count(&count)
		if count > 0 {
			return nil, errors.New("规则名称已存在")
		}
		updates["name"] = req.Name
	}
	if req.NormalLimitHours > 0 {
		updates["normal_limit"] = time.Duration(req.NormalLimitHours) * time.Hour
	}
	if req.SevereLimitHours > 0 {
		updates["severe_limit"] = time.Duration(req.SevereLimitHours) * time.Hour
	}

	if err := database.DB.Model(&rule).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新查询
	database.DB.First(&rule, id)
	return &rule, nil
}

// SetActiveTimeoutRule 设置生效的超时规则
func (s *TimeoutRuleService) SetActiveTimeoutRule(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查规则是否存在
		var rule models.TimeoutRule
		if err := tx.First(&rule, id).Error; err != nil {
			return errors.New("规则不存在")
		}

		// 将所有规则设为不生效
		if err := tx.Model(&models.TimeoutRule{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
			return err
		}

		// 将指定规则设为生效
		if err := tx.Model(&rule).Update("is_active", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// SetInactiveTimeoutRule 禁用超时规则
func (s *TimeoutRuleService) SetInactiveTimeoutRule(id uint) error {
	var rule models.TimeoutRule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return errors.New("规则不存在")
	}

	if !rule.IsActive {
		return errors.New("规则已经是禁用状态")
	}

	return database.DB.Model(&rule).Update("is_active", false).Error
}

// DeleteTimeoutRule 删除超时规则
func (s *TimeoutRuleService) DeleteTimeoutRule(id uint) error {
	var rule models.TimeoutRule
	if err := database.DB.First(&rule, id).Error; err != nil {
		return errors.New("规则不存在")
	}

	if rule.IsActive {
		return errors.New("无法删除生效中的规则")
	}

	return database.DB.Delete(&rule).Error
}

// CreateTimeoutRuleRequest 创建超时规则请求
type CreateTimeoutRuleRequest struct {
	Name             string `json:"name" binding:"required"`
	NormalLimitHours int    `json:"normal_limit_hours" binding:"required,min=1"`
	SevereLimitHours int    `json:"severe_limit_hours" binding:"required,min=1"`
}

// UpdateTimeoutRuleRequest 更新超时规则请求
type UpdateTimeoutRuleRequest struct {
	Name             string `json:"name"`
	NormalLimitHours int    `json:"normal_limit_hours"`
	SevereLimitHours int    `json:"severe_limit_hours"`
}
