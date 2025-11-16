package handler

import (
	"context"
	"strconv"

	"RemindGo/internal/middleware"
	"RemindGo/internal/model"
	"RemindGo/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type TodoHandler struct {
	todoService *service.TodoService
}

// NewTodoHandler 创建待办事项处理器
func NewTodoHandler(todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

// GetTodoList 获取待办事项列表
func (h *TodoHandler) GetTodoList(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	// 解析查询参数
	var params model.TodoQueryParams
	if err := c.Bind(&params); err != nil {
		// 忽略绑定错误，使用默认值
	}

	// 调用service层获取列表
	listResponse, err := h.todoService.GetTodoList(userID, &params)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "获取成功",
		Data:   listResponse,
	})
}

// CreateTodo 创建待办事项
func (h *TodoHandler) CreateTodo(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	var req model.CreateTodoRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "请求参数错误: " + err.Error(),
			Data:   nil,
		})
		return
	}

	// 调用service层创建
	todo, err := h.todoService.CreateTodo(userID, &req)
	if err != nil {
		status := consts.StatusInternalServerError
		if err.Error() == "截止时间格式错误，请使用ISO 8601格式" {
			status = consts.StatusBadRequest
		}
		c.JSON(status, model.BaseResponse{
			Status: status,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusCreated, model.BaseResponse{
		Status: consts.StatusCreated,
		Msg:    "创建成功",
		Data:   todoToResponse(todo),
	})
}

// GetTodo 获取单个待办事项
func (h *TodoHandler) GetTodo(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	id := c.Param("id")
	todoID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "无效的ID",
			Data:   nil,
		})
		return
	}

	// 调用service层获取
	todo, err := h.todoService.GetTodoByID(userID, todoID)
	if err != nil {
		status := consts.StatusInternalServerError
		if err.Error() == "待办事项不存在" {
			status = consts.StatusNotFound
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
		Msg:    "获取成功",
		Data:   todoToResponse(todo),
	})
}

// UpdateTodo 更新待办事项
func (h *TodoHandler) UpdateTodo(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	id := c.Param("id")
	todoID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "无效的ID",
			Data:   nil,
		})
		return
	}

	var req model.UpdateTodoRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "请求参数错误: " + err.Error(),
			Data:   nil,
		})
		return
	}

	// 调用service层更新
	todo, err := h.todoService.UpdateTodo(userID, todoID, &req)
	if err != nil {
		status := consts.StatusInternalServerError
		if err.Error() == "待办事项不存在" {
			status = consts.StatusNotFound
		} else if err.Error() == "截止时间格式错误，请使用ISO 8601格式" {
			status = consts.StatusBadRequest
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
		Data:   todoToResponse(todo),
	})
}

// DeleteTodo 删除待办事项
func (h *TodoHandler) DeleteTodo(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	id := c.Param("id")
	todoID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "无效的ID",
			Data:   nil,
		})
		return
	}

	// 调用service层删除
	if err := h.todoService.DeleteTodo(userID, todoID); err != nil {
		status := consts.StatusInternalServerError
		if err.Error() == "待办事项不存在" {
			status = consts.StatusNotFound
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
		Msg:    "删除成功",
		Data:   nil,
	})
}

// ToggleTodo 切换待办事项状态
func (h *TodoHandler) ToggleTodo(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	id := c.Param("id")
	todoID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(consts.StatusBadRequest, model.BaseResponse{
			Status: consts.StatusBadRequest,
			Msg:    "无效的ID",
			Data:   nil,
		})
		return
	}

	// 调用service层切换状态
	todo, err := h.todoService.ToggleTodo(userID, todoID)
	if err != nil {
		status := consts.StatusInternalServerError
		if err.Error() == "待办事项不存在" {
			status = consts.StatusNotFound
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
		Msg:    "切换成功",
		Data:   todoToResponse(todo),
	})
}

// BatchComplete 批量完成所有待办事项
func (h *TodoHandler) BatchComplete(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	// 调用service层批量完成
	count, err := h.todoService.BatchComplete(userID)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量完成成功",
		Data: model.BatchOperationResult{
			AffectedCount: count,
		},
	})
}

// BatchPending 批量重置所有已完成事项
func (h *TodoHandler) BatchPending(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	// 调用service层批量重置
	count, err := h.todoService.BatchPending(userID)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量重置成功",
		Data: model.BatchOperationResult{
			AffectedCount: count,
		},
	})
}

// BatchClearCompleted 批量删除已完成事项
func (h *TodoHandler) BatchClearCompleted(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	// 调用service层批量删除
	count, err := h.todoService.BatchClearCompleted(userID)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量删除成功",
		Data: model.BatchOperationResult{
			AffectedCount: count,
		},
	})
}

// BatchClearPending 批量删除待办事项
func (h *TodoHandler) BatchClearPending(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	// 调用service层批量删除
	count, err := h.todoService.BatchClearPending(userID)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量删除成功",
		Data: model.BatchOperationResult{
			AffectedCount: count,
		},
	})
}

// GetStats 获取统计信息
func (h *TodoHandler) GetStats(ctx context.Context, c *app.RequestContext) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, model.BaseResponse{
			Status: consts.StatusUnauthorized,
			Msg:    "未认证",
			Data:   nil,
		})
		return
	}

	// 调用service层获取统计
	stats, err := h.todoService.GetStats(userID)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    err.Error(),
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "获取成功",
		Data:   stats,
	})
}

// todoToResponse 将Todo模型转换为响应格式
func todoToResponse(todo *model.Todo) model.TodoResponse {
	resp := model.TodoResponse{
		ID:        todo.ID,
		Title:     todo.Title,
		Content:   todo.Content,
		Status:    todo.Status,
		CreatedAt: todo.CreatedAt.Unix(),
		UpdatedAt: todo.UpdatedAt.Unix(),
	}

	if todo.Deadline != nil {
		deadline := todo.Deadline.Unix()
		resp.Deadline = &deadline
	}

	if todo.CompletedAt != nil {
		completedAt := todo.CompletedAt.Unix()
		resp.CompletedAt = &completedAt
	}

	return resp
}
