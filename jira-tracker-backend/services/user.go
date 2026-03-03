package services

import (
	"errors"
	"fmt"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"
	"jira-tracker-backend/utils"

	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// Login 用户登录
func (s *UserService) Login(phone, code, ip, userAgent string) (*models.User, string, string, error) {
	// 根据配置决定是否验证验证码
	if config.GlobalConfig.Security.EnableCodeVerification {
		smsService := NewSMSService()
		valid, err := smsService.VerifyCode(phone, code)
		if err != nil {
			return nil, "", "", fmt.Errorf("验证码验证失败: %w", err)
		}
		if !valid {
			return nil, "", "", errors.New("验证码错误或已过期")
		}
	}

	// 查找或创建用户
	var user models.User
	var err error
	err = database.DB.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新用户
			user = models.User{
				Phone:  phone,
				Name:   phone, // 默认使用手机号作为姓名
				Role:   "dev", // 默认角色为研发
				Status: 1,
			}
			if err := database.DB.Create(&user).Error; err != nil {
				return nil, "", "", fmt.Errorf("创建用户失败: %w", err)
			}
		} else {
			return nil, "", "", fmt.Errorf("查询用户失败: %w", err)
		}
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, "", "", errors.New("用户账号已被禁用")
	}

	// 生成会话ID
	sessionID := utils.GenerateSessionID()

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Phone, user.Role, sessionID, ip, userAgent)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成token失败: %w", err)
	}

	// 创建会话记录
	sessionService := NewSessionService()
	if err := sessionService.CreateSession(user.ID, token, sessionID, ip, userAgent); err != nil {
		return nil, "", "", fmt.Errorf("创建会话失败: %w", err)
	}

	// 生成CSRF token
	csrfToken, err := utils.GenerateCSRFToken(sessionID)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成CSRF token失败: %w", err)
	}

	return &user, token, csrfToken, nil
}

// CreateUser 创建用户（管理员）
func (s *UserService) CreateUser(req *CreateUserRequest) (*models.User, error) {
	// 检查手机号是否已存在
	var count int64
	database.DB.Model(&models.User{}).Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		return nil, errors.New("该手机号已被注册")
	}

	user := &models.User{
		Phone:  req.Phone,
		Name:   req.Name,
		Role:   req.Role,
		Status: 1,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户（管理员）
func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest) error {
	updates := make(map[string]any)

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return errors.New("没有需要更新的字段")
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// DeleteUser 删除用户（管理员）
func (s *UserService) DeleteUser(userID uint) error {
	// 检查用户是否有待处理工单
	var count int64
	database.DB.Model(&models.Ticket{}).Where("current_user_id = ? AND status IN (?)", userID, []string{"processing", "retesting"}).Count(&count)
	if count > 0 {
		return errors.New("该用户有待处理工单，无法删除")
	}

	if err := database.DB.Delete(&models.User{}, userID).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// GetUserList 获取用户列表（管理员）
func (s *UserService) GetUserList(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 统计总数
	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := database.DB.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// 请求和响应结构体
type CreateUserRequest struct {
	Phone string `json:"phone" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Role  string `json:"role" binding:"required,oneof=admin test dev"`
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Role   string `json:"role" binding:"omitempty,oneof=admin test dev"`
	Status *int   `json:"status" binding:"omitempty,oneof=0 1"`
}
