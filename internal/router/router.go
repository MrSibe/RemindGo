package router

import (
	"RemindGo/internal/handler"
	"RemindGo/internal/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/jwt"
)

func SetupRoutes(h *server.Hertz,
	userHandler *handler.UserHandler,
	todoHandler *handler.TodoHandler,
	jwtMiddleware *jwt.HertzJWTMiddleware) {

	// 引入全局中间件
	h.Use(middleware.CORS())
	h.Use(middleware.Logger())
	h.Use(middleware.Recovery())

	// API v1路由组
	v1 := h.Group("/api/v1")
	{
		// 认证相关路由 (不需要JWT认证)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)                             // 用户注册
			auth.POST("/login", jwtMiddleware.LoginHandler)                          // 用户登录（使用JWT中间件）
			auth.GET("/refresh_token", jwtMiddleware.RefreshHandler)                 // 刷新Token
			auth.POST("/logout", jwtMiddleware.MiddlewareFunc(), userHandler.Logout) // 用户登出（需要认证）
		}

		// 用户相关路由 (需要JWT认证)
		users := v1.Group("/users")
		users.Use(jwtMiddleware.MiddlewareFunc())
		{
			users.GET("/profile", userHandler.GetProfile)    // 获取用户信息
			users.PUT("/profile", userHandler.UpdateProfile) // 更新用户信息
		}

		// 待办事项相关路由 (需要JWT认证)
		todos := v1.Group("/todos")
		todos.Use(jwtMiddleware.MiddlewareFunc())
		{
			// 统计信息 (必须在 /:id 之前)
			todos.GET("/stats", todoHandler.GetStats) // 获取统计信息

			// 批量操作 (必须在 /:id 之前)
			batch := todos.Group("/batch")
			{
				batch.PATCH("/complete", todoHandler.BatchComplete)               // 批量完成所有待办事项
				batch.PATCH("/pending", todoHandler.BatchPending)                 // 批量重置所有已完成事项
				batch.DELETE("/clear-completed", todoHandler.BatchClearCompleted) // 批量删除已完成事项
				batch.DELETE("/clear-pending", todoHandler.BatchClearPending)     // 批量删除待办事项
			}

			// 基础CRUD操作
			todos.GET("", todoHandler.GetTodoList)             // 获取待办事项列表
			todos.POST("", todoHandler.CreateTodo)             // 创建待办事项
			todos.GET("/:id", todoHandler.GetTodo)             // 获取单个待办事项
			todos.PUT("/:id", todoHandler.UpdateTodo)          // 更新待办事项
			todos.DELETE("/:id", todoHandler.DeleteTodo)       // 删除待办事项
			todos.PATCH("/:id/toggle", todoHandler.ToggleTodo) // 切换待办事项状态
		}
	}
}
