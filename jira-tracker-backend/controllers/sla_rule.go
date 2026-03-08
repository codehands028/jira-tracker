package controllers

import (
	"net/http"
	"strconv"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// GetSLARules 获取SLA规则列表
func GetSLARules(c *gin.Context) {
	slaService := services.NewSLARuleService()
	rules, err := slaService.GetSLARules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取SLA规则失败"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// GetSLARule 获取单个SLA规则
func GetSLARule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	slaService := services.NewSLARuleService()
	rule, err := slaService.GetSLARule(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// CreateSLARule 创建SLA规则
func CreateSLARule(c *gin.Context) {
	var req services.CreateSLARuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证严重时限必须大于普通时限
	if req.SevereLimitHours <= req.NormalLimitHours {
		c.JSON(http.StatusBadRequest, gin.H{"error": "严重SLA时限必须大于普通SLA时限"})
		return
	}

	slaService := services.NewSLARuleService()
	rule, err := slaService.CreateSLARule(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建SLA规则失败"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// UpdateSLARule 更新SLA规则
func UpdateSLARule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	var req services.UpdateSLARuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slaService := services.NewSLARuleService()
	rule, err := slaService.UpdateSLARule(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// DeleteSLARule 删除SLA规则
func DeleteSLARule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	slaService := services.NewSLARuleService()
	if err := slaService.DeleteSLARule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ToggleSLARule 启用/禁用SLA规则
func ToggleSLARule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slaService := services.NewSLARuleService()
	if err := slaService.ToggleSLARule(uint(id), req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// GetBatchOperationLogs 获取批量操作日志
func GetBatchOperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userIDStr := c.Query("user_id")
	operationType := c.Query("operation_type")

	var userID uint
	if userIDStr != "" {
		id, _ := strconv.ParseUint(userIDStr, 10, 32)
		userID = uint(id)
	}

	logs, total, err := services.GetBatchOperationLogs(page, pageSize, userID, operationType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取操作日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
