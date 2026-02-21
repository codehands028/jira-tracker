package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiter *rate.Limiter
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
}

var globalLimiter *RateLimiter

// InitRateLimiter 初始化限流器
func InitRateLimiter() {
	globalLimiter = &RateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Minute/100), 100), // 每分钟最多100个请求
		visitors: make(map[string]*rate.Limiter),
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")

		if !exists {
			// 未登录用户限制更严格
			if !globalLimiter.limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
				c.Abort()
				return
			}
		} else {
			// 已登录用户按用户限流
			key := fmt.Sprintf("user:%v", userID)
			globalLimiter.mu.RLock()
			limiter, exists := globalLimiter.visitors[key]
			globalLimiter.mu.RUnlock()
			
			if !exists {
				globalLimiter.mu.Lock()
				// 双重检查，防止并发创建
				limiter, exists = globalLimiter.visitors[key]
				if !exists {
					// 为每个用户创建独立的限流器
					limiter = rate.NewLimiter(rate.Every(time.Minute/50), 50) // 每用户每分钟最多50个请求
					globalLimiter.visitors[key] = limiter
				}
				globalLimiter.mu.Unlock()
			}

			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
