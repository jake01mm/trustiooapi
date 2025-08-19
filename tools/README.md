# Trusioo API 工具集

这个目录包含了用于管理和维护 Trusioo API 的各种工具脚本。

## 目录结构

```
tools/
├── db/              # 数据库相关工具
│   ├── migrate.go   # 数据库迁移
│   ├── test.go      # 数据库连接测试
│   └── verify.go    # 验证数据库表结构
├── admin/           # 管理员相关工具
│   └── check_password.go # 检查和修复管理员密码
├── debug/           # 调试工具
│   └── check_data.go # 检查表中数据统计
├── Makefile         # 便捷命令集合
└── README.md        # 本文档
```

## 使用方法

### 方法1：使用 Makefile（推荐）

```bash
cd tools

# 查看所有可用命令
make help

# 执行数据库迁移
make db-migrate

# 测试数据库连接
make db-test

# 验证数据库表结构
make db-verify

# 检查管理员密码
make admin-check

# 检查表数据统计
make debug-data
```

### 方法2：直接运行 Go 文件

```bash
# 在项目根目录执行
go run tools/db/migrate.go
go run tools/db/test.go
go run tools/db/verify.go
go run tools/admin/check_password.go
go run tools/debug/check_data.go
```

## 工具说明

### 数据库工具 (db/)

- **migrate.go** - 执行数据库迁移，创建所需的表结构
- **test.go** - 测试数据库连接是否正常
- **verify.go** - 验证数据库中的表是否正确创建

### 管理员工具 (admin/)

- **check_password.go** - 检查默认管理员账户的密码哈希，如果不正确会自动修复

### 调试工具 (debug/)

- **check_data.go** - 显示各个表中的数据统计，用于调试和监控

## 使用场景

1. **初次部署**：运行 `make db-migrate` 创建数据库表
2. **连接问题**：运行 `make db-test` 检查数据库连接
3. **结构验证**：运行 `make db-verify` 确认表结构正确
4. **管理员问题**：运行 `make admin-check` 修复管理员登录问题
5. **数据调试**：运行 `make debug-data` 查看数据统计

## 注意事项

- 所有工具都需要正确的 `.env` 配置文件
- 运行前确保数据库服务正常运行
- 工具会自动加载项目配置，无需额外参数