package model

import "time"

type Todo struct {
	ID          int64      `json:"id" gorm:"primary_key"`
	UserID      int64      `json:"-" gorm:"not null;index"`
	Title       string     `json:"title" gorm:"not null;size:255"`
	Content     string     `json:"content" gorm:"type:text"`
	Status      int        `json:"status" gorm:"default:0;index"` // 0-待办，1-已完成
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Deadline    *time.Time `json:"deadline" gorm:"index"`
	CompletedAt *time.Time `json:"completed_at"`
}

// CreateTodoRequest 创建待办事项请求
type CreateTodoRequest struct {
	Title    string `json:"title" binding:"required,min=1,max=255"`
	Content  string `json:"content" binding:"max=1000"`
	Deadline string `json:"deadline"` // ISO 8601 格式
}

// UpdateTodoRequest 更新待办事项请求
type UpdateTodoRequest struct {
	Title    *string `json:"title" binding:"omitempty,min=1,max=255"`
	Content  *string `json:"content" binding:"omitempty,max=1000"`
	Deadline *string `json:"deadline"` // ISO 8601 格式
	Status   *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// TodoResponse 单个待办事项响应
type TodoResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Status      int    `json:"status"`
	CreatedAt   int64  `json:"created_at"`   // Unix 时间戳
	UpdatedAt   int64  `json:"updated_at"`   // Unix 时间戳
	Deadline    *int64 `json:"deadline"`     // Unix 时间戳
	CompletedAt *int64 `json:"completed_at"` // Unix 时间戳
}

// TodoListResponse 待办事项列表响应
type TodoListResponse struct {
	Items      []TodoResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// TodoStats 待办事项统计
type TodoStats struct {
	Total          int64   `json:"total"`
	Pending        int64   `json:"pending"`
	Completed      int64   `json:"completed"`
	CompletionRate float64 `json:"completion_rate"`
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	AffectedCount int64 `json:"affected_count"`
}

// TodoQueryParams 查询参数
type TodoQueryParams struct {
	Status    string `query:"status"`     // all, pending, completed
	Page      int    `query:"page"`       // 页码，从1开始
	PageSize  int    `query:"page_size"`  // 每页条数
	Keyword   string `query:"keyword"`    // 搜索关键词
	SortBy    string `query:"sort_by"`    // 排序字段
	SortOrder string `query:"sort_order"` // 排序方式: asc, desc
}
