package main

import (
	"log"
	"time"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"gorm.io/gorm"
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

	// 自动迁移表结构
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Ticket{},
		&models.TicketFlow{},
		&models.Notification{},
		&models.OperationLog{},
		&models.TimeoutRule{},
		&models.SLARule{},
		&models.BatchOperationLog{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed")

	// 初始化数据
	initTimeoutRules()
	initSLARules()
	initAdminUser()

	log.Println("Data initialization completed")
}

// initTimeoutRules 初始化超时规则
func initTimeoutRules() {
	var count int64
	database.DB.Model(&models.TimeoutRule{}).Count(&count)
	if count > 0 {
		log.Println("Timeout rules already exist, skipping initialization")
		return
	}

	rules := []models.TimeoutRule{
		{
			Name:        "默认规则",
			NormalLimit: 24 * time.Hour,
			SevereLimit: 48 * time.Hour,
			IsActive:    true,
		},
		{
			Name:        "紧急规则",
			NormalLimit: 4 * time.Hour,
			SevereLimit: 8 * time.Hour,
			IsActive:    false,
		},
		{
			Name:        "宽松规则",
			NormalLimit: 48 * time.Hour,
			SevereLimit: 72 * time.Hour,
			IsActive:    false,
		},
	}

	for _, rule := range rules {
		if err := database.DB.Create(&rule).Error; err != nil {
			log.Printf("Failed to create timeout rule: %v", err)
		} else {
			log.Printf("Created timeout rule: %s", rule.Name)
		}
	}
}

// initSLARules 初始化SLA规则
func initSLARules() {
	var count int64
	database.DB.Model(&models.SLARule{}).Count(&count)
	if count > 0 {
		log.Println("SLA rules already exist, skipping initialization")
		return
	}

	rules := []models.SLARule{
		{
			Name:          "紧急工单时限",
			Priority:      "critical",
			NormalLimit:   4 * time.Hour,
			SevereLimit:   8 * time.Hour,
			IsActive:      true,
			PriorityOrder: 10,
		},
		{
			Name:          "高优先级时限",
			Priority:      "high",
			NormalLimit:   8 * time.Hour,
			SevereLimit:   16 * time.Hour,
			IsActive:      true,
			PriorityOrder: 10,
		},
		{
			Name:          "中优先级时限",
			Priority:      "medium",
			NormalLimit:   24 * time.Hour,
			SevereLimit:   48 * time.Hour,
			IsActive:      true,
			PriorityOrder: 10,
		},
		{
			Name:          "低优先级时限",
			Priority:      "low",
			NormalLimit:   48 * time.Hour,
			SevereLimit:   72 * time.Hour,
			IsActive:      true,
			PriorityOrder: 10,
		},
		{
			Name:          "默认时限",
			NormalLimit:   24 * time.Hour,
			SevereLimit:   48 * time.Hour,
			IsActive:      true,
			PriorityOrder: 0,
		},
	}

	for _, rule := range rules {
		if err := database.DB.Create(&rule).Error; err != nil {
			log.Printf("Failed to create SLA rule: %v", err)
		} else {
			log.Printf("Created SLA rule: %s", rule.Name)
		}
	}
}

// initAdminUser 初始化管理员用户
func initAdminUser() {
	// 预定义的管理员账号列表
	adminAccounts := []struct {
		Phone string
		Name  string
	}{
		{"13800138000", "系统管理员"},
	}

	for _, account := range adminAccounts {
		var user models.User
		result := database.DB.Where("phone = ?", account.Phone).First(&user)
		
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新管理员
			admin := models.User{
				Phone:     account.Phone,
				Name:      account.Name,
				Role:      "admin",
				Status:    1,
				TicketNum: 0,
			}
			if err := database.DB.Create(&admin).Error; err != nil {
				log.Printf("Failed to create admin user %s: %v", account.Phone, err)
			} else {
				log.Printf("Created admin user: %s (phone: %s)", admin.Name, admin.Phone)
			}
		} else if result.Error == nil {
			// 用户已存在，检查是否为管理员，不是则更新
			if user.Role != "admin" {
				if err := database.DB.Model(&user).Update("role", "admin").Error; err != nil {
					log.Printf("Failed to update user %s to admin: %v", account.Phone, err)
				} else {
					log.Printf("Updated user %s to admin role", account.Phone)
				}
			} else {
				log.Printf("Admin user already exists: %s (phone: %s)", user.Name, user.Phone)
			}
		} else {
			log.Printf("Error checking admin user: %v", result.Error)
		}
	}

	// 创建测试用户
	testUsers := []models.User{
		{Phone: "13800138001", Name: "张三", Role: "test", Status: 1, TicketNum: 0},
		{Phone: "13800138002", Name: "李四", Role: "dev", Status: 1, TicketNum: 0},
		{Phone: "13800138003", Name: "王五", Role: "dev", Status: 1, TicketNum: 0},
		{Phone: "13800138004", Name: "赵六", Role: "test", Status: 1, TicketNum: 0},
	}

	for _, user := range testUsers {
		var existingUser models.User
		result := database.DB.Where("phone = ?", user.Phone).First(&existingUser)
		if result.Error == gorm.ErrRecordNotFound {
			if err := database.DB.Create(&user).Error; err != nil {
				log.Printf("Failed to create test user: %v", err)
			} else {
				log.Printf("Created test user: %s (%s)", user.Name, user.Role)
			}
		}
	}
}

// CreateTestData 创建测试数据（可选）
func CreateTestData() {
	// 检查是否已有测试数据
	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	if userCount > 1 {
		log.Println("Test data already exists, skipping")
		return
	}

	// 创建测试用户
	testUsers := []models.User{
		{Phone: "13800138001", Name: "张三", Role: "test", Status: 1, TicketNum: 0},
		{Phone: "13800138002", Name: "李四", Role: "dev", Status: 1, TicketNum: 0},
		{Phone: "13800138003", Name: "王五", Role: "dev", Status: 1, TicketNum: 0},
		{Phone: "13800138004", Name: "赵六", Role: "test", Status: 1, TicketNum: 0},
	}

	for _, user := range testUsers {
		if err := database.DB.Create(&user).Error; err != nil {
			log.Printf("Failed to create test user: %v", err)
		} else {
			log.Printf("Created test user: %s (%s)", user.Name, user.Role)
		}
	}

	// 创建测试工单
	testTickets := []models.Ticket{
		{
			JiraKey:       "PROJ-001",
			JiraURL:       "https://jira.example.com/browse/PROJ-001",
			Description:   "登录功能异常，无法正常登录系统",
			Priority:      "high",
			Status:        "processing",
			CurrentUserID: 2, // 李四
			IsTimeout:     false,
		},
		{
			JiraKey:       "PROJ-002",
			JiraURL:       "https://jira.example.com/browse/PROJ-002",
			Description:   "数据导出功能崩溃",
			Priority:      "critical",
			Status:        "processing",
			CurrentUserID: 3, // 王五
			IsTimeout:     true,
			TimeoutLevel:  "normal",
		},
		{
			JiraKey:       "PROJ-003",
			JiraURL:       "https://jira.example.com/browse/PROJ-003",
			Description:   "界面显示错位问题",
			Priority:      "medium",
			Status:        "retesting",
			CurrentUserID: 4, // 赵六
			IsTimeout:     false,
		},
	}

	for _, ticket := range testTickets {
		if err := database.DB.Create(&ticket).Error; err != nil {
			log.Printf("Failed to create test ticket: %v", err)
		} else {
			log.Printf("Created test ticket: %s", ticket.JiraKey)

			// 创建初始流转记录
			flow := models.TicketFlow{
				TicketID:   ticket.ID,
				FromUserID: 1, // 管理员创建
				ToUserID:   ticket.CurrentUserID,
				Content:    "创建工单",
				IsTimeout:  false,
			}
			database.DB.Create(&flow)

			// 更新用户待处理工单数
			database.DB.Model(&models.User{}).Where("id = ?", ticket.CurrentUserID).
				UpdateColumn("ticket_num", gorm.Expr("ticket_num + ?", 1))
		}
	}
}

// 为方便执行，添加一个简单的命令行入口
func init() {
	// 可以通过命令行参数控制是否创建测试数据
	// go run scripts/init_data.go --with-test-data
}
