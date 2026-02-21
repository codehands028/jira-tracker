package router

import (
	"jira-tracker-backend/controllers"
	"jira-tracker-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 使用中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.OperationLogMiddleware())
	r.Use(middleware.RateLimitMiddleware())  // 新增：限流中间件

	// 公开路由
	auth := r.Group("/api/auth")
	{
		auth.POST("/send-code", controllers.SendCode)
		auth.POST("/login", controllers.Login)
		auth.POST("/logout", middleware.AuthMiddleware(), controllers.Logout)  // 新增：登出接口
	}

	// 需要认证的路由
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// 个人信息
		profile := api.Group("/profile")
		{
			profile.GET("", controllers.GetProfile)
			profile.PUT("", controllers.UpdateProfile)
		}

		// 用户管理
		users := api.Group("/users")
		{
			// 获取用户列表（所有认证用户可访问，用于选择处理人）
			users.GET("", controllers.GetUserList)
			users.GET("/:id", controllers.GetUserByID)
		}
		// 用户管理（管理员）
		usersAdmin := api.Group("/users")
		usersAdmin.Use(middleware.AdminMiddleware())
		{
			usersAdmin.POST("", controllers.CreateUser)
			usersAdmin.PUT("/:id", controllers.UpdateUser)
			usersAdmin.DELETE("/:id", controllers.DeleteUser)
		}

		// 工单管理
		tickets := api.Group("/tickets")
		{
			// 创建工单（测试或管理员）
			tickets.POST("", middleware.TestOrAdminMiddleware(), controllers.CreateTicket)
			// 获取工单列表
			tickets.GET("", controllers.GetTicketList)
			// 获取工单详情
			tickets.GET("/:id", controllers.GetTicketDetail)
			// 流转工单
			tickets.POST("/:id/flow", controllers.FlowTicket)
			// 复测工单（测试或管理员）
			tickets.POST("/:id/retest", middleware.TestOrAdminMiddleware(), controllers.RetestTicket)
			// 关闭工单（测试或管理员）
			tickets.POST("/:id/close", middleware.TestOrAdminMiddleware(), controllers.CloseTicket)
			// 批量分配工单（管理员）
			tickets.POST("/batch/assign", middleware.AdminMiddleware(), controllers.BatchAssignTickets)
			// 批量关闭工单（测试或管理员）
			tickets.POST("/batch/close", middleware.TestOrAdminMiddleware(), controllers.BatchCloseTickets)
		}

		// 统计数据
		statistics := api.Group("/statistics")
		{
			statistics.GET("", middleware.AdminMiddleware(), controllers.GetStatistics)
			statistics.GET("/dashboard", controllers.GetDashboard)
		}

		// 通知
		notifications := api.Group("/notifications")
		{
			notifications.GET("", controllers.GetNotifications)
			notifications.PUT("/:id/read", controllers.MarkNotificationAsRead)
		}

		// 操作日志
		logs := api.Group("/logs")
		{
			logs.GET("", middleware.AdminMiddleware(), controllers.GetOperationLogs)
			logs.GET("/my", controllers.GetUserOperationLogs)
		}

		// 超时规则管理（管理员）
		timeoutRules := api.Group("/timeout-rules")
		timeoutRules.Use(middleware.AdminMiddleware())
		{
			timeoutRules.GET("", controllers.GetTimeoutRules)
			timeoutRules.GET("/active", controllers.GetActiveTimeoutRule)
			timeoutRules.POST("", controllers.CreateTimeoutRule)
			timeoutRules.PUT("/:id", controllers.UpdateTimeoutRule)
			timeoutRules.PUT("/:id/active", controllers.SetActiveTimeoutRule)
			timeoutRules.PUT("/:id/inactive", controllers.SetInactiveTimeoutRule)
			timeoutRules.DELETE("/:id", controllers.DeleteTimeoutRule)
		}
	}

	return r
}
