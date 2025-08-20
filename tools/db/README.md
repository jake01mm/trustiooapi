# 数据库工具目录

这个目录包含了 Trusioo API 的数据库管理工具。

## 📁 目录结构

```
db/
├── README.md           # 本说明文件
├── migrate_new.go      # 主迁移工具（推荐使用）
├── status.go          # 迁移状态检查
├── init.go            # 迁移系统初始化
├── backup.go          # 数据库备份和恢复
└── legacy/            # 传统工具（向后兼容）
    ├── migrate/       # 旧迁移工具
    ├── alter/         # 表结构修改
    ├── test/          # 连接测试
    └── verify/        # 结构验证
```

## 🚀 推荐使用（新工具）

### 主要工具

| 文件 | 用途 | 命令 |
|------|------|------|
| `migrate_new.go` | 主迁移工具 | `make db-migrate` |
| `status.go` | 状态检查 | `make db-status` |
| `init.go` | 系统初始化 | `make db-init` |
| `backup.go` | 备份恢复 | `make db-backup` |

### 特性

- ✅ **版本化管理** - 自动追踪迁移版本
- ✅ **回滚支持** - 安全的迁移回滚
- ✅ **状态追踪** - 清晰的迁移状态显示
- ✅ **安全机制** - 备份、确认提示
- ✅ **团队协作** - 防止迁移冲突

## 🔧 传统工具（legacy/）

为了向后兼容保留的工具：

| 目录 | 用途 | 使用场景 |
|------|------|----------|
| `migrate/` | 旧迁移系统 | 紧急情况或特殊需求 |
| `alter/` | 表结构修改 | 单个表的快速修改 |
| `test/` | 连接测试 | 诊断数据库连接问题 |
| `verify/` | 结构验证 | 验证表结构完整性 |

## 📋 使用建议

### 日常开发（推荐）

```bash
# 使用新的迁移系统
make db-status        # 检查状态
make db-create NAME=feature  # 创建迁移
make db-migrate       # 执行迁移
make db-rollback      # 回滚迁移
```

### 特殊情况

```bash
# 使用传统工具
make db-test          # 测试连接
make db-verify        # 验证结构
```

## 🔄 迁移路径

从旧系统迁移到新系统：

1. **检查现状**: `make db-status`
2. **开始使用**: 新迁移使用 `make db-create`
3. **逐步替换**: 旧工具保留用于紧急情况

## 🛡️ 安全提示

- **生产环境**: 执行迁移前先备份 `make db-backup`
- **团队协作**: 拉取代码后检查 `make db-status`
- **测试优先**: 重要迁移先在测试环境验证

---

💡 **建议**: 优先使用新工具，legacy 工具仅用于向后兼容或紧急情况。