package service

import (
	"errors"

	"RemindGo/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register 用户注册
func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	var existingUser model.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("邮箱已被使用")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("创建用户失败")
	}

	return &user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID int64) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID int64, req *model.UpdateUserRequest) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})

	if req.Username != nil {
		// 检查用户名是否已被占用
		var existingUser model.User
		if err := s.db.Where("username = ? AND id != ?", *req.Username, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("用户名已被占用")
		}
		updates["username"] = *req.Username
	}

	if req.Email != nil {
		// 检查邮箱是否已被占用
		var existingUser model.User
		if err := s.db.Where("email = ? AND id != ?", *req.Email, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("邮箱已被占用")
		}
		updates["email"] = *req.Email
	}

	// 执行更新
	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, errors.New("更新失败")
		}
		// 重新查询用户信息
		s.db.First(&user, userID)
	}

	return &user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userID int64) error {
	result := s.db.Delete(&model.User{}, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

// VerifyPassword 验证密码
func (s *UserService) VerifyPassword(user *model.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}
