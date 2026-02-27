# Jira 工单跟踪与效率管理平台 - 快速开始指南

## 📋 项目概述

Jira Tracker 是一个工单跟踪与效率管理平台，提供工单流转管理、超时提醒、数据统计等功能。本项目为学习开发前后端代码的首个项目，功能还不完善，希望得到建议完善项目功能。

### 核心功能
- ✅ 工单录入与流转管理
- ✅ 自动超时检测与提醒
- ✅ 数据看板与统计分析
- ✅ 用户权限管理
- ✅ 操作日志记录
- ✅ 批量操作功能
- ✅ 深色模式支持
- ✅ 多浏览器兼容（Chrome/Firefox/Edge）

---

## 🚀 快速开始

### 1. 环境要求

- **Go**: 1.21 或更高版本
- **Node.js**: 16.0 或更高版本
- **MySQL**: 5.7 或更高版本
- **Redis**: 6.0 或更高版本

### 2. 配置文件说明

#### 2.1 后端配置

编辑 `jira-tracker-backend/config/config.yaml`：
```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: "your_password"  # 修改为你的密码
  database: jira_tracker
  charset: utf8mb4

redis:
  host: localhost
  port: 6379
  password: ""  # 如果有密码，填写密码
  db: 0

server:
  port: 8080  # 后端服务端口
  mode: debug  # 运行模式: debug/release
```

#### 2.2 前端配置

前端访问后端的路径可以通过环境变量文件进行设置：

1. 编辑 `jira-tracker-frontend/.env.development` 文件
2. 配置后端API基础路径：
```bash
VITE_API_BASE_URL=http://localhost:8080/api
```

**注意**：修改配置后需要重启前端服务

#### 2.3 数据库配置

##### 2.3.1 创建数据库
```bash
# 方式一：使用 SQL 脚本
mysql -u root -p < jira-tracker-backend/scripts/init.sql

# 方式二：手动创建
mysql -u root -p
CREATE DATABASE jira_tracker DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 初始化数据

```bash
cd jira-tracker-backend
go run scripts/init_data.go
```

这将自动：
- 创建数据库表结构
- 插入默认超时规则
- 创建管理员账号

**默认管理员账号**：
- 手机号：`13800138000`(短信未对接)
- 验证码：任意6位数字（开发模式下）

### 4. 启动服务

#### 方式一：使用启动脚本（推荐）

**Mac/Linux**：
```bash
chmod +x start.sh
./start.sh
```

**Windows**：
```cmd
start.bat
```

#### 方式二：手动启动

**启动后端**：
```bash
cd jira-tracker-backend
go run main.go
```

**启动前端**：
```bash
cd jira-tracker-frontend
npm install
npm run dev
```

### 5. 访问系统

- **前端地址**：http://localhost:3000
- **后端地址**：http://localhost:8080

### 6. 使用 Nginx 反向代理（生产环境推荐）

项目提供了 Nginx 配置文件，用于在生产环境中使用 Nginx 作为反向代理。

#### 6.1 配置说明

配置文件位于 `nginx/` 目录：
- `nginx.conf` - Nginx 主配置文件
- `jira-tracker.conf` - 项目站点配置文件

#### 6.2 部署步骤

1. **安装 Nginx**
```bash
# Ubuntu/Debian
sudo apt-get install nginx

# CentOS/RHEL
sudo yum install nginx

# macOS
brew install nginx
```

2. **复制配置文件**
```bash
# 复制项目配置到 Nginx 配置目录
sudo cp nginx/jira-tracker.conf /etc/nginx/conf.d/
# 或根据你的 Nginx 配置目录调整路径
```

3. **修改配置文件**

编辑 `/etc/nginx/conf.d/jira-tracker.conf`，根据实际情况修改：
- `server_name` - 你的域名
- `proxy_pass` - 后端服务地址
- `root` - 前端静态文件路径

4. **重启 Nginx**
```bash
# 测试配置文件
sudo nginx -t

# 重启 Nginx
sudo systemctl restart nginx
# 或
sudo service nginx restart
```

5. **验证部署**
访问配置的域名，确认系统正常运行。

---

## 📖 功能模块说明

### 1. 数据看板（Dashboard）
- 工单总览（总数、处理中、待复测、超时）
- 今日统计（新增、关闭、超时）
- 用户工作负载
- 超时预警
- 最近活动

### 2. 工单管理
- 工单列表（支持筛选、分页）
- 工单详情（流转历史时间线）
- 工单流转
- 工单复测
- 工单关闭
- 批量分配（管理员）
- 批量关闭

### 3. 通知中心
- 工单流转通知
- 超时提醒通知(未对接)
- 工单关闭通知
- 未读标记

### 4. 数据统计（管理员）
- 人员维度统计（处理数、平均时长、超时次数）
- 工单维度统计（总时长、流转次数）

### 5. 用户管理（管理员）
- 用户 CRUD
- 角色分配（管理员/测试/研发）
- 状态启用/禁用

### 6. 操作日志（管理员）
- 操作记录查询
- 用户、模块、时间筛选

### 7. 超时规则（管理员）
- 规则 CRUD
- 规则启用/停用
- 普通超时、严重超时配置

### 8. 界面设置
- 深色/浅色模式切换
- 多浏览器兼容（Chrome/Firefox/Edge）

---

## 👥 用户角色权限

| 功能 | 管理员 | 测试 | 研发 |
|------|--------|------|------|
| 数据看板 | ✓ | ✓ | ✓ |
| 工单列表 | ✓ | ✓ | ✓ |
| 创建工单 | ✓ | ✓ | ✗ |
| 流转工单 | ✓ | ✓ | ✓* |
| 复测工单 | ✓ | ✓ | ✗ |
| 关闭工单 | ✓ | ✓ | ✗ |
| 批量分配 | ✓ | ✗ | ✗ |
| 批量关闭 | ✓ | ✓ | ✗ |
| 数据统计 | ✓ | ✗ | ✗ |
| 用户管理 | ✓ | ✗ | ✗ |
| 操作日志 | ✓ | ✗ | ✗ |
| 超时规则 | ✓ | ✗ | ✗ |

*注：研发人员只能流转分配给自己的工单

---

## 🔧 开发说明

### 后端技术栈
- **框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL
- **缓存**：Redis
- **认证**：JWT

### 前端技术栈
- **框架**：Vue 3 (Composition API)
- **UI组件**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router 4
- **HTTP客户端**：Axios
- **构建工具**：Vite

### 目录结构
```
jira-tracker/
├── jira-tracker-backend/          # 后端代码
│   ├── config/                    # 配置
│   ├── controllers/               # 控制器
│   ├── database/                  # 数据库连接
│   ├── middleware/                # 中间件
│   ├── models/                    # 数据模型
│   ├── router/                    # 路由
│   ├── scripts/                   # 脚本
│   ├── services/                  # 业务逻辑
│   └── utils/                     # 工具函数
├── jira-tracker-frontend/         # 前端代码
│   ├── src/
│   │   ├── api/                   # API 接口
│   │   ├── layouts/               # 布局组件
│   │   ├── router/                # 路由配置
│   │   ├── stores/                # 状态管理
│   │   ├── utils/                 # 工具函数
│   │   └── views/                 # 页面组件
│   └── vite.config.js             # Vite 配置
├── nginx/                          # Nginx 配置
│   ├── nginx.conf                 # Nginx 主配置
│   └── jira-tracker.conf          # 项目配置
├── start.sh                        # 启动脚本
└── stop.sh                         # 停止脚本
```

---

## 📝 API 接口文档

### 认证接口
- `POST /api/auth/send-code` - 发送验证码
- `POST /api/auth/login` - 用户登录

### 工单接口
- `GET /api/tickets` - 获取工单列表
- `GET /api/tickets/:id` - 获取工单详情
- `POST /api/tickets` - 创建工单
- `POST /api/tickets/:id/flow` - 流转工单
- `POST /api/tickets/:id/retest` - 复测工单
- `POST /api/tickets/:id/close` - 关闭工单
- `POST /api/tickets/batch/assign` - 批量分配
- `POST /api/tickets/batch/close` - 批量关闭

### 统计接口
- `GET /api/statistics/dashboard` - 获取数据看板
- `GET /api/statistics` - 获取统计数据

### 其他接口
- `GET /api/notifications` - 获取通知列表
- `PUT /api/notifications/:id/read` - 标记已读
- `GET /api/logs` - 获取操作日志（管理员）
- `GET /api/logs/my` - 获取我的操作日志
- `GET /api/timeout-rules` - 获取超时规则（管理员）
- `POST /api/timeout-rules` - 创建超时规则（管理员）
- `PUT /api/timeout-rules/:id` - 更新超时规则（管理员）
- `PUT /api/timeout-rules/:id/active` - 设置生效规则（管理员）
- `DELETE /api/timeout-rules/:id` - 删除超时规则（管理员）

---

## 🐛 故障排查

### 后端无法启动
1. 检查 MySQL 是否运行：`mysql -u root -p`
2. 检查 Redis 是否运行：`redis-cli ping`
3. 检查配置文件是否正确

### 前端无法启动
1. 删除 `node_modules` 文件夹
2. 运行 `npm install`
3. 运行 `npm run dev`

### 登录问题
- 开发模式下，验证码可以是任意6位数字
- 确保后端服务正常运行

### 数据库连接失败
1. 确认 MySQL 服务正在运行
2. 确认数据库已创建
3. 确认配置文件中的连接信息正确

---

## 📞 联系方式

如有问题，请提交 Issue 或联系开发者。

---

**祝您使用愉快！** 🎉
