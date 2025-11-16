package service

import (
	"errors"

	"RemindGo/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db          *gorm.DB
	userService *UserService
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:          db,
		userService: NewUserService(db),
	}
}

// Login 用户登录认证
func (s *AuthService) Login(username, password string) (*model.User, error) {
	// 查找用户（支持用户名或邮箱登录）
	var user model.User
	if err := s.db.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, errors.New("登录失败")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return &user, nil
}

// ValidateUser 验证用户是否存在且密码正确
func (s *AuthService) ValidateUser(username, password string) (*model.User, error) {
	return s.Login(username, password)
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return errors.New("密码修改失败")
	}

	return nil
}

// ResetPassword 重置密码（管理员功能或忘记密码功能）
func (s *AuthService) ResetPassword(userID int64, newPassword string) error {
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	result := s.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", string(hashedPassword))
	if result.Error != nil {
		return errors.New("密码重置失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}

	return nil
}

// CheckUserExists 检查用户是否存在
func (s *AuthService) CheckUserExists(username, email string) (bool, error) {
	var count int64
	if err := s.db.Model(&model.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsUsernameAvailable 检查用户名是否可用
func (s *AuthService) IsUsernameAvailable(username string) (bool, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsEmailAvailable 检查邮箱是否可用
func (s *AuthService) IsEmailAvailable(email string) (bool, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}
