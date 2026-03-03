package middleware

import (
	"net/http"
	"strings"

	"jira-tracker-backend/utils"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware CSRF保护中间件
// 验证请求中的CSRF token是否有效
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于GET、HEAD、OPTIONS、TRACE方法，跳过CSRF验证
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" || method == "TRACE" {
			c.Next()
			return
		}

		// 从请求头中获取CSRF token
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			// 如果请求头中没有，尝试从表单中获取
			csrfToken = c.PostForm("csrf_token")
		}

		// 获取session ID
		sessionID, exists := c.Get("session_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到会话信息"})
			c.Abort()
			return
		}

		// 验证CSRF token
		if err := utils.ValidateCSRFToken(sessionID.(string), csrfToken); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token验证失败: " + err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CSRFTokenMiddleware 为响应添加CSRF token
// 用于在需要CSRF保护的接口中返回新的token
func CSRFTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果请求成功，为响应添加CSRF token
		if len(c.Errors) == 0 && c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			sessionID, exists := c.Get("session_id")
			if exists {
				// 刷新CSRF token
				newToken, err := utils.RefreshCSRFToken(sessionID.(string))
				if err == nil {
					// 将新的token添加到响应头
					c.Header("X-CSRF-Token", newToken)

					// 如果响应是JSON，也尝试添加到响应体中
					if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
						// 获取响应数据
						response, exists := c.Get("response")
						if exists {
							if respMap, ok := response.(map[string]interface{}); ok {
								respMap["csrf_token"] = newToken
							}
						}
					}
				}
			}
		}
	}
}
