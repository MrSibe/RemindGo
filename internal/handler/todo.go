package handler

import (
	"context"
	"math"
	"strconv"
	"time"

	"RemindGo/internal/middleware"
	"RemindGo/internal/model"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"
)

type TodoHandler struct {
	db *gorm.DB
}

// NewTodoHandler 创建待办事项处理器
func NewTodoHandler(db *gorm.DB) *TodoHandler {
	return &TodoHandler{db: db}
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

	// 设置默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.Status == "" {
		params.Status = "all"
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// 构建查询
	query := h.db.Model(&model.Todo{}).Where("user_id = ?", userID)

	// 状态过滤
	if params.Status == "pending" {
		query = query.Where("status = ?", 0)
	} else if params.Status == "completed" {
		query = query.Where("status = ?", 1)
	}

	// 关键词搜索
	if params.Keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "查询失败",
			Data:   nil,
		})
		return
	}

	// 排序
	orderClause := params.SortBy + " " + params.SortOrder
	query = query.Order(orderClause)

	// 分页
	offset := (params.Page - 1) * params.PageSize
	var todos []model.Todo
	if err := query.Offset(offset).Limit(params.PageSize).Find(&todos).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "查询失败",
			Data:   nil,
		})
		return
	}

	// 转换为响应格式
	items := make([]model.TodoResponse, len(todos))
	for i, todo := range todos {
		items[i] = todoToResponse(&todo)
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "获取成功",
		Data: model.TodoListResponse{
			Items:      items,
			Total:      total,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: totalPages,
		},
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

	todo := model.Todo{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Status:  0, // 默认待办
	}

	// 解析截止时间
	if req.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			c.JSON(consts.StatusBadRequest, model.BaseResponse{
				Status: consts.StatusBadRequest,
				Msg:    "截止时间格式错误，请使用ISO 8601格式",
				Data:   nil,
			})
			return
		}
		todo.Deadline = &deadline
	}

	if err := h.db.Create(&todo).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "创建失败",
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusCreated, model.BaseResponse{
		Status: consts.StatusCreated,
		Msg:    "创建成功",
		Data:   todoToResponse(&todo),
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

	var todo model.Todo
	if err := h.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(consts.StatusNotFound, model.BaseResponse{
				Status: consts.StatusNotFound,
				Msg:    "待办事项不存在",
				Data:   nil,
			})
		} else {
			c.JSON(consts.StatusInternalServerError, model.BaseResponse{
				Status: consts.StatusInternalServerError,
				Msg:    "查询失败",
				Data:   nil,
			})
		}
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "获取成功",
		Data:   todoToResponse(&todo),
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

	var todo model.Todo
	if err := h.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(consts.StatusNotFound, model.BaseResponse{
				Status: consts.StatusNotFound,
				Msg:    "待办事项不存在",
				Data:   nil,
			})
		} else {
			c.JSON(consts.StatusInternalServerError, model.BaseResponse{
				Status: consts.StatusInternalServerError,
				Msg:    "查询失败",
				Data:   nil,
			})
		}
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == 1 {
			now := time.Now()
			updates["completed_at"] = &now
		} else {
			updates["completed_at"] = nil
		}
	}
	if req.Deadline != nil {
		if *req.Deadline == "" {
			updates["deadline"] = nil
		} else {
			deadline, err := time.Parse(time.RFC3339, *req.Deadline)
			if err != nil {
				c.JSON(consts.StatusBadRequest, model.BaseResponse{
					Status: consts.StatusBadRequest,
					Msg:    "截止时间格式错误，请使用ISO 8601格式",
					Data:   nil,
				})
				return
			}
			updates["deadline"] = &deadline
		}
	}

	if len(updates) > 0 {
		if err := h.db.Model(&todo).Updates(updates).Error; err != nil {
			c.JSON(consts.StatusInternalServerError, model.BaseResponse{
				Status: consts.StatusInternalServerError,
				Msg:    "更新失败",
				Data:   nil,
			})
			return
		}
	}

	// 重新查询
	h.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo)

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "更新成功",
		Data:   todoToResponse(&todo),
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

	result := h.db.Where("id = ? AND user_id = ?", todoID, userID).Delete(&model.Todo{})
	if result.Error != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "删除失败",
			Data:   nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(consts.StatusNotFound, model.BaseResponse{
			Status: consts.StatusNotFound,
			Msg:    "待办事项不存在",
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

	var todo model.Todo
	if err := h.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(consts.StatusNotFound, model.BaseResponse{
				Status: consts.StatusNotFound,
				Msg:    "待办事项不存在",
				Data:   nil,
			})
		} else {
			c.JSON(consts.StatusInternalServerError, model.BaseResponse{
				Status: consts.StatusInternalServerError,
				Msg:    "查询失败",
				Data:   nil,
			})
		}
		return
	}

	// 切换状态
	updates := make(map[string]interface{})
	if todo.Status == 0 {
		updates["status"] = 1
		now := time.Now()
		updates["completed_at"] = &now
	} else {
		updates["status"] = 0
		updates["completed_at"] = nil
	}

	if err := h.db.Model(&todo).Updates(updates).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "更新失败",
			Data:   nil,
		})
		return
	}

	// 重新查询
	h.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo)

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "切换成功",
		Data:   todoToResponse(&todo),
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

	now := time.Now()
	result := h.db.Model(&model.Todo{}).
		Where("user_id = ? AND status = ?", userID, 0).
		Updates(map[string]interface{}{
			"status":       1,
			"completed_at": &now,
		})

	if result.Error != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "批量操作失败",
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量完成成功",
		Data: model.BatchOperationResult{
			AffectedCount: result.RowsAffected,
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

	result := h.db.Model(&model.Todo{}).
		Where("user_id = ? AND status = ?", userID, 1).
		Updates(map[string]interface{}{
			"status":       0,
			"completed_at": nil,
		})

	if result.Error != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "批量操作失败",
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量重置成功",
		Data: model.BatchOperationResult{
			AffectedCount: result.RowsAffected,
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

	result := h.db.Where("user_id = ? AND status = ?", userID, 1).Delete(&model.Todo{})

	if result.Error != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "批量删除失败",
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量删除成功",
		Data: model.BatchOperationResult{
			AffectedCount: result.RowsAffected,
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

	result := h.db.Where("user_id = ? AND status = ?", userID, 0).Delete(&model.Todo{})

	if result.Error != nil {
		c.JSON(consts.StatusInternalServerError, model.BaseResponse{
			Status: consts.StatusInternalServerError,
			Msg:    "批量删除失败",
			Data:   nil,
		})
		return
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "批量删除成功",
		Data: model.BatchOperationResult{
			AffectedCount: result.RowsAffected,
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

	var total int64
	var completed int64
	var pending int64

	// 统计总数
	h.db.Model(&model.Todo{}).Where("user_id = ?", userID).Count(&total)

	// 统计已完成
	h.db.Model(&model.Todo{}).Where("user_id = ? AND status = ?", userID, 1).Count(&completed)

	// 统计待办
	h.db.Model(&model.Todo{}).Where("user_id = ? AND status = ?", userID, 0).Count(&pending)

	// 计算完成率
	var completionRate float64
	if total > 0 {
		completionRate = float64(completed) / float64(total)
	}

	c.JSON(consts.StatusOK, model.BaseResponse{
		Status: consts.StatusOK,
		Msg:    "获取成功",
		Data: model.TodoStats{
			Total:          total,
			Pending:        pending,
			Completed:      completed,
			CompletionRate: completionRate,
		},
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
