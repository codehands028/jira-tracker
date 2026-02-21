package controllers

import (
	"net/http"
	"strings"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// SendCode 发送验证码
func SendCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	smsService := services.NewSMSService()
	if err := smsService.SendCode(req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userService := services.NewUserService()

	// 获取客户端信息
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	user, token, csrfToken, err := userService.Login(req.Phone, req.Code, clientIP, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"token":      token,
		"csrf_token": csrfToken,
	})
}

// Logout 用户登出
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供认证令牌"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		// 使会话失效
		sessionService := services.NewSessionService()
		if err := sessionService.InvalidateSession(parts[1]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "登出失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的认证令牌"})
	}
}
