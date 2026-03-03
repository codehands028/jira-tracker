package middleware

import (
	"net/http"
	"strings"

	"jira-tracker-backend/config"
	"jira-tracker-backend/services"
	"jira-tracker-backend/utils"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查请求来源是否在允许列表中
		allowed := false
		for _, allowedOrigin := range config.GlobalConfig.Security.AllowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		// 暴露X-CSRF-Token响应头给前端
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-CSRF-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 解析token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式错误"})
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 验证会话是否有效
		sessionService := services.NewSessionService()
		isValid, err := sessionService.ValidateSession(parts[1])
		if err != nil || !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话已过期或失效，请重新登录"})
			c.Abort()
			return
		}

		// 验证请求来源（仅验证User-Agent，允许IP变化）
		userAgent := c.GetHeader("User-Agent")
		if claims.UserAgent != userAgent {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求来源验证失败，请重新登录"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("role", claims.Role)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TestOrAdminMiddleware 测试或管理员权限中间件
func TestOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || (role != "admin" && role != "test") {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要测试或管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
