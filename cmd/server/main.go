package main

import (
	"context"
	"log"

	"RemindGo/internal/database"
	"RemindGo/internal/handler"
	"RemindGo/internal/middleware"
	"RemindGo/internal/router"
	"RemindGo/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	// 初始化数据库
	// 注意：这里使用的是示例配置，生产环境应该从配置文件或环境变量读取
	dbConfig := database.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "123456",
		DBName:   "remind_go",
	}

	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Service层
	userService := service.NewUserService(db)
	todoService := service.NewTodoService(db)

	// 初始化Handler层
	userHandler := handler.NewUserHandler(userService)
	todoHandler := handler.NewTodoHandler(todoService)

	// 初始化JWT中间件
	jwtMiddleware, err := middleware.NewJWTMiddleware(db)
	if err != nil {
		log.Fatalf("Failed to initialize JWT middleware: %v", err)
	}

	// 初始化Hertz服务器
	h := server.Default(server.WithHostPorts(":8080"))

	// 健康检查接口
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{
			"status": 200,
			"msg":    "pong",
			"data":   nil,
		})
	})

	// 设置路由
	router.SetupRoutes(h, userHandler, todoHandler, jwtMiddleware)

	// 启动服务器
	log.Println("Server is starting on :8080...")
	h.Spin()
}
