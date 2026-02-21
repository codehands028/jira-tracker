package services

import (
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"
	"strconv"
)

type OperationLogService struct{}

func NewOperationLogService() *OperationLogService {
	return &OperationLogService{}
}

// GetOperationLogs 获取操作日志列表
func (s *OperationLogService) GetOperationLogs(page, pageSize int, userID, module, action, startTime, endTime string) ([]models.OperationLog, int64, error) {
	var logs []models.OperationLog
	var total int64

	query := database.DB.Model(&models.OperationLog{})

	if userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			query = query.Where("user_id = ?", id)
		}
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
