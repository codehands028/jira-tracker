package controllers

import (
	"net/http"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// GetProfile 获取个人信息
func GetProfile(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	profileService := services.NewProfileService()
	profile, err := profileService.GetProfile(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取个人信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}

// UpdateProfile 更新个人信息
func UpdateProfile(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profileService := services.NewProfileService()
	if err := profileService.UpdateProfile(userID.(uint), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新个人信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}
