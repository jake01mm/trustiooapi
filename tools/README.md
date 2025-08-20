# Trusioo API 工具集

这个目录包含了 Trusioo API 项目的各种开发和维护工具。

## 🛠️ 可用工具

### 数据库迁移工具 (推荐使用)

基于 `golang-migrate` 的现代化迁移系统，专为 Neon Database 优化：

```bash
# 初始化迁移系统
make db-init

# 查看迁移状态
make db-status

# 执行迁移
make db-migrate

# 回滚迁移
make db-rollback

# 创建新迁移
make db-create NAME=your_migration_name

# 查看所有命令
make help
```

**详细文档**: [../docs/MIGRATIONS.md](../docs/MIGRATIONS.md)

### 传统工具 (向后兼容)

为了向后兼容保留的旧工具（位于 `db/legacy/` 目录）：

- `db/legacy/migrate/migrate.go` - 旧的迁移工具
- `db/legacy/alter/alter.go` - 表结构修改工具
- `db/legacy/test/test.go` - 数据库连接测试
- `db/legacy/verify/verify.go` - 表结构验证

## 🚀 快速开始

### 1. 首次使用

```bash
cd tools
make db-init
```

### 2. 日常开发

```bash
# 检查当前状态
make db-status

# 创建新功能的迁移
make db-create NAME=add_user_preferences

# 编辑生成的迁移文件，然后执行
make db-migrate
```

### 3. 安全操作

```bash
# 执行重要迁移前创建备份
make db-backup

# 如果需要回滚
make db-rollback
```

## 📁 文件结构

```
tools/
├── Makefile              # 主要命令入口
├── README.md            # 本文件
├── db/                  # 数据库工具
│   ├── README.md        # 数据库工具说明
│   ├── migrate_new.go   # 主迁移工具 (推荐)
│   ├── status.go        # 状态检查
│   ├── init.go          # 初始化工具
│   ├── backup.go        # 备份工具
│   └── legacy/          # 传统工具 (向后兼容)
│       ├── migrate/     # 旧迁移工具
│       ├── alter/       # 表修改工具
│       ├── test/        # 连接测试
│       └── verify/      # 结构验证
├── admin/               # 管理员工具
│   └── check_password.go
└── debug/               # 调试工具
    └── check_data.go
```

## 🔄 迁移系统对比

### 新系统 vs 旧系统

| 特性 | 新系统 (golang-migrate) | 旧系统 |
|------|-------------------------|--------|
| 版本控制 | ✅ 自动版本追踪 | ❌ 手动管理 |
| 回滚功能 | ✅ 自动回滚脚本 | ❌ 需要手动编写 |
| 状态追踪 | ✅ 详细状态显示 | ❌ 无状态显示 |
| 安全机制 | ✅ 备份+确认提示 | ❌ 直接执行 |
| 并发保护 | ✅ 防止同时执行 | ❌ 无保护 |
| 幂等性 | ✅ 可重复执行 | ⚠️ 部分支持 |

### 迁移建议

**新项目**: 直接使用新的迁移系统

**现有项目**: 
1. 先使用 `make db-status` 检查当前状态
2. 如果数据库已存在，系统会自动检测并标记为最新版本
3. 后续使用新系统创建迁移

## 🛡️ 最佳实践

### 1. 开发流程

```bash
# 1. 开始新功能开发前
make db-status

# 2. 如需数据库变更，创建迁移
make db-create NAME=add_feature_x

# 3. 编辑迁移文件
# 4. 执行迁移
make db-migrate

# 5. 测试功能
# 6. 如有问题，可快速回滚
make db-rollback
```

### 2. 团队协作

```bash
# 拉取代码后检查是否有新迁移
make db-status

# 如有待处理迁移，执行它们
make db-migrate
```

### 3. 部署流程

```bash
# 生产部署前
make db-backup          # 备份数据库
make db-migrate         # 执行迁移
# 部署应用
# 验证功能正常
```

## 🚨 故障排除

### 常见问题

1. **迁移失败导致脏状态**
   ```bash
   make db-status  # 查看状态
   make db-force N=<version>  # 强制修复
   ```

2. **连接失败**
   - 检查 `.env` 文件中的数据库配置
   - 确认 Neon Database 连接信息正确

3. **版本冲突**
   - 检查 `migrations/` 目录下的文件
   - 确保版本号连续

### 获取帮助

```bash
# 查看所有可用命令
make help

# 查看详细文档
cat ../docs/MIGRATIONS.md
```

## 🔗 相关链接

- [项目主页](../README.md)
- [迁移系统详细文档](../docs/MIGRATIONS.md)
- [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
- [Neon Database 文档](https://neon.tech/docs)

---

**注意**: 建议优先使用新的迁移系统 (`make db-*` 命令)，旧工具仅用于特殊情况或向后兼容。