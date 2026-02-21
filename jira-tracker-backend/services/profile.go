package services

import (
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"gorm.io/gorm"
)

type ProfileService struct {
	db *gorm.DB
}

func NewProfileService() *ProfileService {
	return &ProfileService{
		db: database.DB,
	}
}

// ProfileResponse 个人信息响应
type ProfileResponse struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	Status int    `json:"status"`
}

// GetProfile 获取个人信息
func (s *ProfileService) GetProfile(userID uint) (*ProfileResponse, error) {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &ProfileResponse{
		ID:     user.ID,
		Name:   user.Name,
		Phone:  user.Phone,
		Role:   user.Role,
		Status: user.Status,
	}, nil
}

// UpdateProfile 更新个人信息
func (s *ProfileService) UpdateProfile(userID uint, name string) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("name", name).Error
}
