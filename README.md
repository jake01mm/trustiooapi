# Trusioo API

基于 cardaegis_api 项目结构创建的简化版 API 项目，专注于认证和管理员功能。

## 功能特性

- 用户认证系统 (注册、登录、JWT令牌)
- 管理员认证系统 (登录、用户管理)
- 基础的用户管理功能
- PostgreSQL 数据库支持
- RESTful API 设计

## 项目结构

```
trusioo_api/
├── cmd/                    # 应用程序入口
│   └── main.go
├── config/                 # 配置文件
│   └── config.go
├── internal/               # 内部业务逻辑
│   ├── auth/              # 用户认证模块
│   ├── admin/             # 管理员模块
│   ├── common/            # 公共组件
│   ├── middleware/        # 中间件
│   └── router/            # 路由配置
├── pkg/                   # 可重用的包
│   ├── auth/             # JWT 认证工具
│   ├── database/         # 数据库连接
│   └── logger/           # 日志工具
├── scripts/              # 数据库脚本
│   └── init_db.sql      # 数据库初始化脚本
├── docs/                 # 文档目录
├── static/              # 静态文件
├── tools/               # 管理和维护工具
│   ├── db/             # 数据库工具
│   ├── admin/          # 管理员工具
│   ├── debug/          # 调试工具
│   └── README.md       # 工具使用说明
├── .env.example         # 环境变量示例
├── .gitignore          # Git 忽略文件
├── go.mod              # Go 模块文件
└── README.md           # 项目说明
```

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd trusioo_api
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，设置数据库连接信息等
```

### 4. 初始化数据库

创建 PostgreSQL 数据库，然后执行初始化脚本：

```bash
psql -U postgres -d trusioo_db -f scripts/init_db.sql
```

### 5. 运行项目

```bash
go run cmd/main.go
```

服务将在 `http://localhost:8080` 启动。

## 管理工具

项目提供了丰富的管理和维护工具，位于 `tools/` 目录：

```bash
cd tools

# 查看所有可用工具
make help

# 数据库相关工具
make db-migrate    # 执行数据库迁移
make db-test       # 测试数据库连接
make db-verify     # 验证表结构

# 管理员工具
make admin-check   # 检查管理员密码

# 调试工具
make debug-data    # 检查数据统计
```

详细说明请参考：[tools/README.md](tools/README.md)

## API 接口

### 用户认证

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `GET /api/v1/auth/profile` - 获取用户资料 (需要认证)

### 管理员

- `POST /api/v1/admin/auth/login` - 管理员登录
- `POST /api/v1/admin/auth/refresh` - 刷新管理员令牌
- `GET /api/v1/admin/profile` - 获取管理员资料 (需要认证)
- `GET /api/v1/admin/users/stats` - 获取用户统计 (需要管理员认证)
- `GET /api/v1/admin/users` - 获取用户列表 (需要管理员认证)
- `GET /api/v1/admin/users/{id}` - 获取用户详情 (需要管理员认证)

### 健康检查

- `GET /health` - 健康检查

## 默认管理员账户

- 邮箱: `admin@trusioo.com`
- 密码: `admin123`

## 开发说明

### 数据库表结构

- `users` - 用户表
- `admins` - 管理员表
- `user_refresh_tokens` - 用户刷新令牌表
- `admin_refresh_tokens` - 管理员刷新令牌表
- `user_login_sessions` - 用户登录会话表
- `admin_login_sessions` - 管理员登录会话表
- `verifications` - 验证码表 (预留)

### 认证机制

- 使用 JWT 进行身份认证
- 访问令牌默认有效期 2 小时
- 刷新令牌默认有效期 7 天
- 支持用户和管理员分离的认证体系

### 环境变量

主要环境变量说明：

- `DB_HOST` - 数据库主机
- `DB_PORT` - 数据库端口
- `DB_USER` - 数据库用户名
- `DB_PASSWORD` - 数据库密码
- `DB_NAME` - 数据库名称
- `JWT_SECRET` - JWT 访问令牌密钥
- `JWT_REFRESH_SECRET` - JWT 刷新令牌密钥
- `PORT` - 服务端口

## License

MIT License