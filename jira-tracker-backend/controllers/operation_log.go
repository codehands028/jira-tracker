package controllers

import (
	"net/http"
	"strconv"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// GetOperationLogs 获取操作日志列表
func GetOperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID := c.Query("user_id")
	module := c.Query("module")
	action := c.Query("action")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	logService := services.NewOperationLogService()
	logs, total, err := logService.GetOperationLogs(page, pageSize, userID, module, action, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  logs,
		"total": total,
	})
}

// GetUserOperationLogs 获取当前用户的操作日志
func GetUserOperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	logService := services.NewOperationLogService()
	logs, total, err := logService.GetOperationLogs(page, pageSize, strconv.FormatUint(uint64(userID.(uint)), 10), "", "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  logs,
		"total": total,
	})
}
