package handler

import (
	"context"

	"RemindGo/internal/middleware"
	"RemindGo/internal/model"
	"RemindGo/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
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

	// 调用service层注册用户
	user, err := h.userService.Register(&req)
	if err != nil {
		// 根据错误类型返回不同状态码
		status := consts.StatusInternalServerError
		if err.Error() == "用户名已存在" || err.Error() == "邮箱已被使用" {
			status = consts.StatusConflict
		}
		c.JSON(status, model.BaseResponse{
			Status: status,
			Msg:    err.Error(),
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

	// 调用service层获取用户信息
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(consts.StatusNotFound, model.BaseResponse{
			Status: consts.StatusNotFound,
			Msg:    err.Error(),
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

	// 调用service层更新用户信息
	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		status := consts.StatusInternalServerError
		if err.Error() == "用户不存在" {
			status = consts.StatusNotFound
		} else if err.Error() == "用户名已被占用" || err.Error() == "邮箱已被占用" {
			status = consts.StatusConflict
		}
		c.JSON(status, model.BaseResponse{
			Status: status,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

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
