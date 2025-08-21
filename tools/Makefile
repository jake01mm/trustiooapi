# Trusioo API Tools
# 数据库和管理工具集合

.PHONY: help db-init db-migrate db-rollback db-status db-create db-force db-drop db-backup db-restore db-test db-verify admin-check debug-data

help:
	@echo "=== Trusioo API 数据库工具 ==="
	@echo "迁移管理："
	@echo "  db-init          - 初始化数据库和迁移系统"
	@echo "  db-migrate       - 执行所有待处理的迁移"
	@echo "  db-rollback      - 回滚最后一个迁移"
	@echo "  db-rollback-n N  - 回滚N个迁移"
	@echo "  db-status        - 显示当前迁移版本状态"
	@echo "  db-create NAME   - 创建新的迁移文件"
	@echo "  db-force N       - 强制设置迁移版本（危险操作）"
	@echo "  db-drop          - 删除所有表（危险操作）"
	@echo "  db-backup        - 创建数据库备份"
	@echo "  db-restore FILE  - 从备份恢复（危险操作）"
	@echo ""
	@echo "其他工具："
	@echo "  db-test          - 测试数据库连接"
	@echo "  db-verify        - 验证数据库表结构"
	@echo "  admin-check      - 检查和修复管理员密码"
	@echo "  debug-data       - 检查表中的数据统计"
	@echo ""
	@echo "示例："
	@echo "  make db-create NAME=add_user_preferences"
	@echo "  make db-rollback-n N=2"
	@echo "  make db-force N=1"

db-init:
	@echo "🔧 初始化数据库和迁移系统..."
	@go run db/init/main.go

db-migrate:
	@echo "🚀 执行数据库迁移..."
	@go run db/migrate/main.go -action=up

db-rollback:
	@echo "⬇️  回滚最后一个迁移..."
	@go run db/migrate/main.go -action=down -steps=1

db-rollback-n:
	@echo "⬇️  回滚 $(N) 个迁移..."
	@go run db/migrate/main.go -action=down -steps=$(N)

db-status:
	@go run db/status/main.go

db-create:
	@echo "📝 创建新迁移: $(NAME)"
	@go run db/migrate/main.go -action=create -name="$(NAME)"

db-force:
	@echo "⚡ 强制设置迁移版本到 $(N)..."
	@go run db/migrate/main.go -action=force -force=$(N)

db-drop:
	@echo "🗑️  删除所有表..."
	@read -p "确定要删除所有表吗？这将清空整个数据库！输入 'yes' 确认: " confirm && [ "$$confirm" = "yes" ]
	@go run db/migrate/main.go -action=drop

db-backup:
	@echo "📦 创建数据库备份..."
	@go run db/backup/main.go -action=backup

db-restore:
	@echo "🔄 从备份恢复数据库..."
	@go run db/backup/main.go -action=restore -file="$(FILE)"

db-test:
	@echo "测试数据库连接..."
	@go run db/legacy/test/test.go

db-verify:
	@echo "验证数据库表结构..."
	@go run db/legacy/verify/verify.go

admin-check:
	@echo "检查管理员账户..."
	@go run admin/check_password.go

debug-data:
	@echo "检查表数据..."
	@go run debug/check_data/main.go