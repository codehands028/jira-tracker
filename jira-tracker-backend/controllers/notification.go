package controllers

import (
	"net/http"
	"strconv"

	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"github.com/gin-gonic/gin"
)

// GetNotifications 获取通知列表
func GetNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var notifications []models.Notification
	var total int64

	query := database.DB.Model(&models.Notification{}).Where("user_id = ?", userID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Ticket").Offset(offset).Limit(pageSize).Order("created_at desc").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      notifications,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// MarkNotificationAsRead 标记通知为已读
func MarkNotificationAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	userID, _ := c.Get("user_id")

	// 验证通知是否属于当前用户
	var notification models.Notification
	if err := database.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	if notification.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该通知"})
		return
	}

	// 标记为已读
	if err := database.DB.Model(&notification).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标记成功"})
}
