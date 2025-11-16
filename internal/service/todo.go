package service

import (
	"errors"
	"math"
	"time"

	"RemindGo/internal/model"

	"gorm.io/gorm"
)

// TodoService 待办事项服务
type TodoService struct {
	db *gorm.DB
}

// NewTodoService 创建待办事项服务
func NewTodoService(db *gorm.DB) *TodoService {
	return &TodoService{db: db}
}

// GetTodoList 获取待办事项列表
func (s *TodoService) GetTodoList(userID int64, params *model.TodoQueryParams) (*model.TodoListResponse, error) {
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
	query := s.db.Model(&model.Todo{}).Where("user_id = ?", userID)

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
		return nil, errors.New("查询失败")
	}

	// 排序
	orderClause := params.SortBy + " " + params.SortOrder
	query = query.Order(orderClause)

	// 分页
	offset := (params.Page - 1) * params.PageSize
	var todos []model.Todo
	if err := query.Offset(offset).Limit(params.PageSize).Find(&todos).Error; err != nil {
		return nil, errors.New("查询失败")
	}

	// 转换为响应格式
	items := make([]model.TodoResponse, len(todos))
	for i, todo := range todos {
		items[i] = s.todoToResponse(&todo)
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	return &model.TodoListResponse{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CreateTodo 创建待办事项
func (s *TodoService) CreateTodo(userID int64, req *model.CreateTodoRequest) (*model.Todo, error) {
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
			return nil, errors.New("截止时间格式错误，请使用ISO 8601格式")
		}
		todo.Deadline = &deadline
	}

	if err := s.db.Create(&todo).Error; err != nil {
		return nil, errors.New("创建失败")
	}

	return &todo, nil
}

// GetTodoByID 获取单个待办事项
func (s *TodoService) GetTodoByID(userID, todoID int64) (*model.Todo, error) {
	var todo model.Todo
	if err := s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("待办事项不存在")
		}
		return nil, errors.New("查询失败")
	}
	return &todo, nil
}

// UpdateTodo 更新待办事项
func (s *TodoService) UpdateTodo(userID, todoID int64, req *model.UpdateTodoRequest) (*model.Todo, error) {
	var todo model.Todo
	if err := s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("待办事项不存在")
		}
		return nil, errors.New("查询失败")
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
				return nil, errors.New("截止时间格式错误，请使用ISO 8601格式")
			}
			updates["deadline"] = &deadline
		}
	}

	if len(updates) > 0 {
		if err := s.db.Model(&todo).Updates(updates).Error; err != nil {
			return nil, errors.New("更新失败")
		}
		// 重新查询
		s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo)
	}

	return &todo, nil
}

// DeleteTodo 删除待办事项
func (s *TodoService) DeleteTodo(userID, todoID int64) error {
	result := s.db.Where("id = ? AND user_id = ?", todoID, userID).Delete(&model.Todo{})
	if result.Error != nil {
		return errors.New("删除失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("待办事项不存在")
	}
	return nil
}

// ToggleTodo 切换待办事项状态
func (s *TodoService) ToggleTodo(userID, todoID int64) (*model.Todo, error) {
	var todo model.Todo
	if err := s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("待办事项不存在")
		}
		return nil, errors.New("查询失败")
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

	if err := s.db.Model(&todo).Updates(updates).Error; err != nil {
		return nil, errors.New("更新失败")
	}

	// 重新查询
	s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo)
	return &todo, nil
}

// BatchComplete 批量完成所有待办事项
func (s *TodoService) BatchComplete(userID int64) (int64, error) {
	now := time.Now()
	result := s.db.Model(&model.Todo{}).
		Where("user_id = ? AND status = ?", userID, 0).
		Updates(map[string]interface{}{
			"status":       1,
			"completed_at": &now,
		})

	if result.Error != nil {
		return 0, errors.New("批量操作失败")
	}
	return result.RowsAffected, nil
}

// BatchPending 批量重置所有已完成事项
func (s *TodoService) BatchPending(userID int64) (int64, error) {
	result := s.db.Model(&model.Todo{}).
		Where("user_id = ? AND status = ?", userID, 1).
		Updates(map[string]interface{}{
			"status":       0,
			"completed_at": nil,
		})

	if result.Error != nil {
		return 0, errors.New("批量操作失败")
	}
	return result.RowsAffected, nil
}

// BatchClearCompleted 批量删除已完成事项
func (s *TodoService) BatchClearCompleted(userID int64) (int64, error) {
	result := s.db.Where("user_id = ? AND status = ?", userID, 1).Delete(&model.Todo{})
	if result.Error != nil {
		return 0, errors.New("批量删除失败")
	}
	return result.RowsAffected, nil
}

// BatchClearPending 批量删除待办事项
func (s *TodoService) BatchClearPending(userID int64) (int64, error) {
	result := s.db.Where("user_id = ? AND status = ?", userID, 0).Delete(&model.Todo{})
	if result.Error != nil {
		return 0, errors.New("批量删除失败")
	}
	return result.RowsAffected, nil
}

// GetStats 获取统计信息
func (s *TodoService) GetStats(userID int64) (*model.TodoStats, error) {
	var total int64
	var completed int64
	var pending int64

	// 统计总数
	s.db.Model(&model.Todo{}).Where("user_id = ?", userID).Count(&total)

	// 统计已完成
	s.db.Model(&model.Todo{}).Where("user_id = ? AND status = ?", userID, 1).Count(&completed)

	// 统计待办
	s.db.Model(&model.Todo{}).Where("user_id = ? AND status = ?", userID, 0).Count(&pending)

	// 计算完成率
	var completionRate float64
	if total > 0 {
		completionRate = float64(completed) / float64(total)
	}

	return &model.TodoStats{
		Total:          total,
		Pending:        pending,
		Completed:      completed,
		CompletionRate: completionRate,
	}, nil
}

// todoToResponse 将待办事项模型转换为响应格式
func (s *TodoService) todoToResponse(todo *model.Todo) model.TodoResponse {
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
