# 备忘录API服务

基于Golang Hertz/Gin框架的RESTful API备忘录服务，帮助FanOne管理寒假待办事项，实现弯道超车！

## 项目特色

- 🔐 **JWT认证** - 安全的用户认证机制
- 📝 **RESTful API** - 符合REST设计规范的接口
- 🗄️ **三层架构** - 清晰的项目结构设计
- 🔍 **搜索分页** - 支持关键词搜索和分页查询
- 📊 **统计功能** - 待办事项完成率统计
- 🔄 **批量操作** - 支持批量更新和删除
- 🛡️ **安全防护** - SQL注入防护和数据验证

## 技术栈

- **框架**: Hertz (CloudWeGo) / Gin
- **数据库**: MySQL 8.0+
- **缓存**: Redis (可选)
- **认证**: JWT (github.com/golang-jwt/jwt)
- **ORM**: GORM
- **文档**: OpenAPI 3.0
- **测试**: Postman / Apifox

## 项目结构

```
memo-app/
├── cmd/                    # 应用入口
│   └── main.go            # 主程序入口
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   │   └── config.go     # 配置结构体
│   ├── handler/          # 处理器层（控制器）
│   │   ├── auth.go       # 认证处理器
│   │   ├── todo.go       # 待办事项处理器
│   │   └── user.go       # 用户处理器
│   ├── middleware/       # 中间件
│   │   ├── auth.go       # JWT认证中间件
│   │   ├── cors.go       # CORS中间件
│   │   └── logger.go     # 日志中间件
│   ├── model/            # 数据模型层
│   │   ├── user.go       # 用户模型
│   │   └── todo.go       # 待办事项模型
│   ├── router/           # 路由层
│   │   └── router.go     # 路由配置
│   └── service/          # 业务逻辑层
│       ├── auth.go       # 认证服务
│       ├── todo.go       # 待办事项服务
│       └── user.go       # 用户服务
├── pkg/                  # 公共包
│   ├── database/         # 数据库连接
│   │   └── mysql.go      # MySQL连接
│   ├── jwt/              # JWT工具
│   │   └── jwt.go        # JWT生成和验证
│   ├── response/         # 响应封装
│   │   └── response.go   # 统一响应格式
│   └── validator/        # 验证器
│       └── validator.go  # 参数验证
├── docs/                 # 文档
│   ├── openapi.yaml      # OpenAPI文档
│   └── database.sql      # 数据库脚本
├── config/               # 配置文件
│   ├── app.yaml          # 应用配置
│   └── config.yaml       # 环境配置
├── scripts/              # 脚本文件
│   └── init-db.sh        # 数据库初始化脚本
├── go.mod                # Go模块文件
├── go.sum                # Go依赖校验
├── Makefile              # 构建脚本
├── Dockerfile            # Docker配置
├── .gitignore           # Git忽略文件
└── README.md            # 项目说明
```

## 三层架构设计

### 1. 表现层 (Handler)
- **职责**: 处理HTTP请求，参数验证，返回响应
- **位置**: `internal/handler/`
- **特点**: 不包含业务逻辑，只负责请求转发

### 2. 业务逻辑层 (Service)
- **职责**: 处理核心业务逻辑
- **位置**: `internal/service/`
- **特点**: 包含业务规则，调用数据访问层

### 3. 数据访问层 (Model)
- **职责**: 数据模型定义，数据库操作
- **位置**: `internal/model/`
- **特点**: 使用ORM进行数据库交互

## 核心功能模块

### 🔐 用户认证模块
- 用户注册（用户名、邮箱唯一性验证）
- 用户登录（JWT令牌生成）
- 用户信息管理
- 密码加密存储

### 📝 待办事项模块
- **创建**: 添加新的待办事项
- **查询**: 支持分页、状态过滤、关键词搜索
- **更新**: 修改事项内容和状态
- **删除**: 支持单个和批量删除
- **统计**: 完成率统计和数据分析

### 📊 数据统计功能
- 总事项数统计
- 待办/已完成分类统计
- 完成率计算
- 时间趋势分析

### 🔍 高级查询功能
- 多条件组合查询
- 关键词全文搜索
- 时间范围过滤
- 排序和分页

## API接口概览

### 认证接口
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出

### 用户接口
- `GET /api/v1/users/profile` - 获取用户信息
- `PUT /api/v1/users/profile` - 更新用户信息

### 待办事项接口
- `GET /api/v1/todos` - 获取待办事项列表
- `POST /api/v1/todos` - 创建待办事项
- `GET /api/v1/todos/{id}` - 获取单个事项
- `PUT /api/v1/todos/{id}` - 更新事项
- `DELETE /api/v1/todos/{id}` - 删除事项
- `PATCH /api/v1/todos/{id}/toggle` - 切换状态

### 批量操作接口
- `PATCH /api/v1/todos/batch/complete` - 批量完成
- `PATCH /api/v1/todos/batch/pending` - 批量重置
- `DELETE /api/v1/todos/batch/clear-completed` - 清除已完成
- `DELETE /api/v1/todos/batch/clear-pending` - 清除待办

### 统计接口
- `GET /api/v1/todos/stats` - 获取统计信息

## 数据模型设计

### 用户表 (users)
```sql
- id: 主键，自增
- username: 用户名，唯一
- email: 邮箱，唯一
- password_hash: 密码哈希
- created_at: 创建时间
- updated_at: 更新时间
```

### 待办事项表 (todos)
```sql
- id: 主键，自增
- user_id: 用户ID，外键
- title: 标题
- content: 内容
- status: 状态（0-待办，1-已完成）
- created_at: 创建时间
- updated_at: 更新时间
- deadline: 截止时间
- completed_at: 完成时间
```

## 安全设计

### 🔐 认证安全
- JWT令牌认证机制
- 密码bcrypt加密存储
- 令牌过期时间控制
- Refresh Token机制（可选）

### 🛡️ 数据安全
- SQL注入防护（ORM参数化查询）
- XSS攻击防护（输入过滤）
- 数据权限验证（用户隔离）
- 参数验证和 sanitization

### 📝 接口安全
- 请求频率限制
- IP白名单/黑名单
- CORS跨域控制
- HTTPS传输加密

## 部署说明

### 环境要求
- Go 1.19+
- MySQL 8.0+
- Redis 6.0+（可选）

### 快速开始
```bash
# 克隆项目
git clone https://github.com/yourusername/memo-app.git
cd memo-app

# 安装依赖
go mod download

# 配置数据库
cp config/app.example.yaml config/app.yaml
# 编辑配置文件

# 运行数据库迁移
make migrate

# 启动应用
make run
```

### Docker部署
```bash
# 构建镜像
docker build -t memo-app .

# 运行容器
docker-compose up -d
```

## 测试指南

### 单元测试
```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./internal/service

# 生成测试覆盖率
go test -cover ./...
```

### API测试
- 使用Postman导入OpenAPI文档
- 使用Apifox进行接口测试
- 提供测试用例集合

## 性能优化

### 数据库优化
- 索引优化（用户ID、状态、创建时间）
- 分页查询优化
- 连接池配置
- 读写分离（高并发场景）

### 缓存策略
- Redis缓存热点数据
- JWT令牌缓存
- 查询结果缓存
- 缓存失效策略

### API优化
- 响应数据压缩
- 请求频率限制
- 接口响应时间监控
- 慢查询优化

## 监控运维

### 日志管理
- 结构化日志记录
- 日志级别控制
- 日志轮转和归档
- 错误日志告警

### 性能监控
- API响应时间监控
- 数据库性能监控
- 系统资源监控
- 业务指标监控

## 扩展功能

### 🔮 未来规划
- 🔔 消息通知（邮件/短信提醒）
- 📱 移动端支持
- 🤖 AI智能推荐
- 📊 数据可视化
- 🔄 数据导入导出
- 👥 协作功能

### 🎯 Bonus功能
- 自动生成接口文档
- 三层架构设计模式
- 数据库交互安全性
- 优秀的返回结构设计
- Redis缓存集成

## 开发规范

### 代码规范
- 遵循Go语言编码规范
- 使用gofmt格式化代码
- 完整的注释和文档
- 错误处理规范

### Git规范
- 清晰的提交信息
- 分支管理策略
- 代码审查流程
- 版本标签管理

## 贡献指南

欢迎提交Issue和Pull Request来改进项目！

### 开发流程
1. Fork项目
2. 创建特性分支
3. 提交更改
4. 推送分支
5. 创建Pull Request

## 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 📧 Email: support@memoapp.com
- 🐛 Issue: [GitHub Issues](https://github.com/yourusername/memo-app/issues)
- 📖 Wiki: [项目Wiki](https://github.com/yourusername/memo-app/wiki)

---

**让FanOne的寒假不再摸鱼，实现弯道超车！🚀**