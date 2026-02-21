package database

import (
	"fmt"
	"log"
	"time"

	"jira-tracker-backend/config"
	"jira-tracker-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	cfg := config.GlobalConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(config.GlobalConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.GlobalConfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.GlobalConfig.Database.ConnMaxLifetime)

	// 自动迁移表结构
	err = DB.AutoMigrate(
		&models.User{},
		&models.Ticket{},
		&models.TicketFlow{},
		&models.Notification{},
		&models.OperationLog{},
		&models.TimeoutRule{},
		&models.Session{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化默认超时规则
	initDefaultTimeoutRules()

	log.Println("Database initialized successfully")
	return nil
}

func initDefaultTimeoutRules() {
	var count int64
	// 检查是否存在任何超时规则（包括已删除的）
	DB.Unscoped().Model(&models.TimeoutRule{}).Count(&count)
	if count == 0 {
		// 只有当数据库中没有任何超时规则时，才创建默认规则
		defaultRule := &models.TimeoutRule{
			Name:        "default",
			NormalLimit: 24 * time.Hour,
			SevereLimit: 48 * time.Hour,
			IsActive:    true,
		}
		if err := DB.Create(defaultRule).Error; err != nil {
			log.Printf("Failed to create default timeout rule: %v", err)
		}
	}
}
