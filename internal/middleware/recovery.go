package middleware

import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Recovery 恢复中间件
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(consts.StatusInternalServerError, map[string]interface{}{
					"status": consts.StatusInternalServerError,
					"msg":    "服务器内部错误",
					"data":   nil,
				})
			}
		}()
		c.Next(ctx)
	}
}
