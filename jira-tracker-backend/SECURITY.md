# 安全增强功能说明

本文档说明了系统新增的安全功能，用于防止接口被抓包后恶意请求的情况。

## 1. 会话管理

### 功能描述
- 为每个登录会话创建唯一的会话记录
- 验证请求用户是否为当前登录用户
- 支持会话失效和清理过期会话

### 实现文件
- `models/session.go` - 会话数据模型
- `services/session.go` - 会话服务层
- `utils/session.go` - 会话ID生成工具

### 数据库表结构
```go
type Session struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null;index:idx_user_id"`
    Token     string    `gorm:"type:varchar(500);uniqueIndex;not null"`
    IP        string    `gorm:"type:varchar(50)"`
    UserAgent string    `gorm:"type:varchar(500)"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    ExpiresAt time.Time `gorm:"index:idx_expires_at"`
    IsActive  bool      `gorm:"default:true;index:idx_is_active"`
}
```

### 主要功能
1. **CreateSession** - 创建会话记录
   - 在用户登录时创建
   - 记录用户ID、token、IP地址和User-Agent
   - 设置过期时间（与JWT过期时间一致）

2. **ValidateSession** - 验证会话有效性
   - 验证token是否存在且有效
   - 检查会话是否过期
   - 检查会话是否被禁用

3. **InvalidateSession** - 使会话失效
   - 在用户登出时调用
   - 将会话标记为不活跃

4. **CleanupExpiredSessions** - 清理过期会话
   - 定时任务（每小时）清理过期会话
   - 释放数据库资源

## 2. 请求来源验证

### 功能描述
- 验证请求IP地址是否与登录时一致
- 验证请求User-Agent是否与登录时一致
- 防止token被复制到其他设备使用

### 实现位置
- `middleware/auth.go` - AuthMiddleware中间件
- `utils/jwt.go` - JWT Claims增强

### 主要功能
1. **JWT Claims增强**
   ```go
   type Claims struct {
       UserID    uint
       Phone     string
       Role      string
       SessionID string  // 新增：会话ID
       IP        string  // 新增：登录时的IP地址
       UserAgent string  // 新增：登录时的User-Agent
       jwt.RegisteredClaims
   }
   ```

2. **AuthMiddleware验证**
   - 解析JWT token获取用户信息
   - 验证会话是否有效
   - 验证请求IP地址
   - 验证请求User-Agent
   - 验证失败返回401错误

## 3. 请求频率限制

### 功能描述
- 限制单个用户的请求频率
- 防止暴力破解和恶意请求
- 未登录用户限制更严格

### 实现文件
- `middleware/rate_limit.go` - 限流中间件

### 限制规则
1. **未登录用户**
   - 每分钟最多100个请求
   - 全局限制

2. **已登录用户**
   - 每用户每分钟最多50个请求
   - 按用户独立限制

### 主要功能
```go
type RateLimiter struct {
    limiter *rate.Limiter
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
}
```

### 使用方式
- 在`router.go`中全局启用限流中间件
- 自动区分已登录和未登录用户
- 超出限制返回429状态码

## 4. CSRF保护

### 功能描述
- 生成CSRF Token
- 防止跨站请求伪造攻击

### 实现文件
- `utils/csrf.go` - CSRF工具类

### 主要功能
1. **GenerateCSRFToken**
   - 基于会话ID和时间生成token
   - 使用SHA256哈希算法

2. **ValidateCSRFToken**
   - 验证CSRF token有效性
   - 可扩展为Redis存储验证

### 使用方式
- 在登录时生成CSRF token
- 在响应中返回给前端
- 前端在需要CSRF保护的请求中携带

## 5. 登出功能

### 功能描述
- 使用户当前会话失效
- 清理会话记录
- 防止token被重复使用

### 实现位置
- `controllers/auth.go` - Logout控制器
- `router/router.go` - 登出路由

### 主要功能
```go
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
    }
}
```

## 6. 定时任务

### 会话清理任务
- **执行频率**：每小时
- **功能**：清理过期会话
- **实现**：`startSessionCleaner`函数

### 超时检查任务
- **执行频率**：每5分钟
- **功能**：检查工单超时
- **实现**：`startTimeoutChecker`函数

## 7. 接口变更

### 新增接口
1. **POST /api/auth/logout**
   - 功能：用户登出
   - 认证：需要
   - 说明：使当前会话失效

### 修改接口
1. **POST /api/auth/login**
   - 变更：返回CSRF token
   - 变更：创建会话记录
   - 变更：记录客户端信息

## 8. 安全建议

### 前端建议
1. **存储CSRF token**
   - 登录成功后存储CSRF token
   - 在需要CSRF保护的请求中携带token

2. **处理登出**
   - 提供登出按钮
   - 清除本地存储的token
   - 跳转到登录页面

3. **处理认证错误**
   - 捕获401错误
   - 显示友好的错误提示
   - 引导用户重新登录

### 后端建议
1. **监控异常行为**
   - 记录频繁的认证失败
   - 记录异常的请求来源
   - 设置告警阈值

2. **定期审计**
   - 审查活跃会话
   - 审查异常登录行为
   - 审查可疑的请求模式

3. **日志记录**
   - 记录所有认证失败
   - 记录会话创建和失效
   - 记录限流触发事件

## 9. 部署说明

### 数据库迁移
系统启动时会自动创建Session表，无需手动执行迁移。

### 环境变量
确保以下环境变量正确配置：
- JWT_EXPIRE：JWT过期时间
- 数据库连接信息

### 初始化
系统启动时会自动：
1. 初始化限流器
2. 启动会话清理任务
3. 启动超时检查任务

## 10. 测试建议

### 功能测试
1. 测试正常登录流程
2. 测试会话验证
3. 测试登出功能
4. 测试请求频率限制
5. 测试请求来源验证

### 安全测试
1. 尝试使用已失效的token
2. 尝试在不同IP使用同一token
3. 尝试暴力破解登录
4. 尝试高频请求接口

## 11. 注意事项

1. **会话管理**
   - 确保数据库有足够的存储空间
   - 定期检查会话表大小
   - 考虑使用Redis存储会话以提高性能

2. **限流策略**
   - 根据实际业务调整限流阈值
   - 监控限流触发频率
   - 考虑实现IP黑名单机制

3. **监控告警**
   - 监控异常认证行为
   - 监控高频请求来源
   - 设置合理的告警阈值

4. **性能优化**
   - 会话验证使用缓存
   - 限流器使用Redis实现
   - 考虑读写分离
