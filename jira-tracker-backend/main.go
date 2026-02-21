package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"
	"jira-tracker-backend/middleware"
	"jira-tracker-backend/router"
	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis
	if err := database.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}

	// 初始化限流器
	middleware.InitRateLimiter()

	// 设置Gin运行模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 初始化路由
	r := router.SetupRouter()

	// 启动超时检查定时任务
	go startTimeoutChecker()

	// 启动会话清理定时任务
	go startSessionCleaner()

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GlobalConfig.Server.Port),
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("Server is running on port %s", config.GlobalConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 设置5秒超时以完成正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

// startTimeoutChecker 启动超时检查定时任务
func startTimeoutChecker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ticketService := services.NewTicketService()
		if err := ticketService.CheckTimeout(); err != nil {
			log.Printf("Failed to check timeout: %v", err)
		}
	}
}

// startSessionCleaner 启动会话清理定时任务
func startSessionCleaner() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sessionService := services.NewSessionService()
		if err := sessionService.CleanupExpiredSessions(); err != nil {
			log.Printf("Failed to cleanup expired sessions: %v", err)
		}
	}
}
