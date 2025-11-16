package handler

import (
	"context"

	"RemindGo/internal/middleware"
	"RemindGo/internal/model"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// Register 用户注册
func (h *UserHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "请求参数错误: " + err.Error(),
			Data:   nil,
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(consts.StatusConflict, model.BaseResponse{
			Status: consts.StatusConflict,
			Msg:    "用户名已存在",
			Data:   nil,
		})
		return
	}

	// 检查邮箱是否已存在
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(consts.StatusConflict, model.BaseResponse{
			Status: consts.StatusConflict,
			Msg:    "邮箱已被使用",
			Data:   nil,
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "服务器错误",
			Data:   nil,
		})
		return
	}

	// 创建用户
	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "创建用户失败",
			Data:   nil,
		})
		return
	}

	// 返回响应（注册成功，但不自动登录，需要用户手动登录）
	c.JSON(consts.StatusCreated, model.BaseResponse{
		Status: consts.StatusCreated,
		Msg:    "注册成功，请登录",
		Data: model.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
		},
	})
}

// Login 用户登录 - 注意：登录逻辑现在由JWT中间件的LoginHandler处理
// 这个方法保留用于向后兼容或自定义登录逻辑

// Logout 用户登出
func (h *UserHandler) Logout(ctx context.Context, c *app.RequestContext) {
	// 在无状态JWT的实现中，登出通常在客户端完成（删除token）
	// 如果需要服务端控制，可以实现token黑名单机制
	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "登出成功",
		Data:   nil,
	})
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(consts.StatusNotFound, model.BaseResponse{
			Status: consts.StatusNotFound,
			Msg:    "用户不存在",
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "获取成功",
		Data: model.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
		},
	})
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	var req model.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "请求参数错误: " + err.Error(),
			Data:   nil,
		})
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(consts.StatusNotFound, model.BaseResponse{
			Status: consts.StatusNotFound,
			Msg:    "用户不存在",
			Data:   nil,
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Username != nil {
		// 检查用户名是否已被占用
		var existingUser model.User
		if err := h.db.Where("username = ? AND id != ?", *req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(consts.StatusConflict, model.BaseResponse{
				Status: consts.StatusConflict,
				Msg:    "用户名已被占用",
				Data:   nil,
			})
			return
		}
		updates["username"] = *req.Username
	}

	if req.Email != nil {
		// 检查邮箱是否已被占用
		var existingUser model.User
		if err := h.db.Where("email = ? AND id != ?", *req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(consts.StatusConflict, model.BaseResponse{
				Status: consts.StatusConflict,
				Msg:    "邮箱已被占用",
				Data:   nil,
			})
			return
		}
		updates["email"] = *req.Email
	}

	// 执行更新
	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(consts.StatusInternalServerError, model.BaseResponse{
				Status: consts.StatusInternalServerError,
				Msg:    "更新失败",
				Data:   nil,
			})
			return
		}
	}

	// 重新查询用户信息
	h.db.First(&user, userID)

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "更新成功",
		Data: model.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
		},
	})
}
