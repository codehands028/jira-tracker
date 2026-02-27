package middleware

import (
	"strings"

	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志中间件
func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只记录需要认证的请求
		if userID, exists := c.Get("user_id"); exists {
			// 不记录查询操作（GET请求）
			if c.Request.Method == "GET" {
				return
			}

			// 检查用户是否存在
			var userExists bool
			database.DB.Model(&models.User{}).Where("id = ?", userID).Select("count(*) > 0").Scan(&userExists)

			// 只有用户存在时才记录操作日志
			if userExists {
				log := &models.OperationLog{
					UserID:      userID.(uint),
					Module:      getModule(c.Request.URL.Path),
					Action:      c.Request.Method,
					Description: c.Request.URL.Path,
					IPAddress:   c.ClientIP(),
				}

				database.DB.Create(log)
			}
		}
	}
}

func getModule(path string) string {
	// 简单的模块识别逻辑
	switch {
	case strings.HasPrefix(path, "/api/auth"):
		return "auth"
	case strings.HasPrefix(path, "/api/users"):
		return "user"
	case strings.HasPrefix(path, "/api/profile"):
		return "user"
	case strings.HasPrefix(path, "/api/tickets"):
		return "ticket"
	case strings.HasPrefix(path, "/api/notifications"):
		return "notification"
	case strings.HasPrefix(path, "/api/statistics"):
		return "statistics"
	case strings.HasPrefix(path, "/api/logs"):
		return "operation_log"
	case strings.HasPrefix(path, "/api/timeout-rules"):
		return "timeout_rule"
	default:
		return "unknown"
	}
}
