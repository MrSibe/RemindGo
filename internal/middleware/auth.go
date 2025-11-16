package middleware

import (
	"context"
	"errors"
	"time"

	"RemindGo/internal/model"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// JWTSecret JWT密钥，生产环境应该从配置文件或环境变量读取
	JWTSecret = []byte("your-secret-key-change-this-in-production")
	// TokenExpiration Token过期时间
	TokenExpiration = time.Hour * 24 * 7 // 7天
	// identityKey 用于在上下文中存储用户信息的键
	identityKey = "user_id"
)

// JWTUser JWT用户信息
type JWTUser struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// NewJWTMiddleware 创建JWT中间件
func NewJWTMiddleware(db *gorm.DB) (*jwt.HertzJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.HertzJWTMiddleware{
		Realm:       "RemindGo",
		Key:         JWTSecret,
		Timeout:     TokenExpiration,
		MaxRefresh:  TokenExpiration,
		IdentityKey: identityKey,

		// PayloadFunc 定义JWT中存储的数据
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*JWTUser); ok {
				return jwt.MapClaims{
					identityKey: v.UserID,
					"username":  v.Username,
				}
			}
			return jwt.MapClaims{}
		},

		// IdentityHandler 从JWT中提取用户信息
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			userID, _ := claims[identityKey].(float64) // JWT会将数字转为float64
			username, _ := claims["username"].(string)
			return &JWTUser{
				UserID:   int64(userID),
				Username: username,
			}
		},

		// Authenticator 用户认证逻辑
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var req model.LoginRequest
			if err := c.BindAndValidate(&req); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			// 查找用户（支持用户名或邮箱登录）
			var user model.User
			if err := db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			// 验证密码
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &JWTUser{
				UserID:   user.ID,
				Username: user.Username,
			}, nil
		},

		// LoginResponse 自定义登录响应
		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			// 从claims中获取用户信息
			claims := jwt.ExtractClaims(ctx, c)
			userID, _ := claims[identityKey].(float64)

			// 从数据库获取完整用户信息
			var dbUser model.User
			db.First(&dbUser, int64(userID))

			c.JSON(consts.StatusOK, model.BaseResponse{
				Status: consts.StatusOK,
				Msg:    "登录成功",
				Data: model.LoginResponse{
					Token: token,
					User: model.UserInfo{
						ID:        dbUser.ID,
						Username:  dbUser.Username,
						Email:     dbUser.Email,
						CreatedAt: dbUser.CreatedAt.Unix(),
					},
				},
			})
		},

		// Unauthorized 未授权响应
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(code, model.BaseResponse{
				Status: code,
				Msg:    message,
				Data:   nil,
			})
		},

		// TokenLookup 从哪里查找token
		TokenLookup: "header: Authorization",

		// TokenHeadName token前缀
		TokenHeadName: "Bearer",

		// TimeFunc 时间函数
		TimeFunc: time.Now,
	})

	return authMiddleware, err
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *app.RequestContext) (int64, error) {
	user, exists := c.Get(identityKey)
	if !exists {
		// 尝试从JWT_PAYLOAD获取
		if jwtUser, ok := c.Get("JWT_PAYLOAD"); ok {
			if u, ok := jwtUser.(*JWTUser); ok {
				return u.UserID, nil
			}
		}
		return 0, errors.New("未找到用户ID")
	}

	// 处理从IdentityHandler返回的用户信息
	if jwtUser, ok := user.(*JWTUser); ok {
		return jwtUser.UserID, nil
	}

	return 0, errors.New("未找到用户ID")
}

// GetUsername 从上下文获取用户名
func GetUsername(c *app.RequestContext) (string, error) {
	user, exists := c.Get(identityKey)
	if !exists {
		// 尝试从JWT_PAYLOAD获取
		if jwtUser, ok := c.Get("JWT_PAYLOAD"); ok {
			if u, ok := jwtUser.(*JWTUser); ok {
				return u.Username, nil
			}
		}
		return "", errors.New("未找到用户名")
	}

	// 处理从IdentityHandler返回的用户信息
	if jwtUser, ok := user.(*JWTUser); ok {
		return jwtUser.Username, nil
	}

	return "", errors.New("未找到用户名")
}
