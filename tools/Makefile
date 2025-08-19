# Trusioo API Tools
# 数据库和管理工具集合

.PHONY: help db-migrate db-test db-verify admin-check debug-data

help:
	@echo "可用的工具命令："
	@echo "  db-migrate    - 执行数据库迁移"
	@echo "  db-test       - 测试数据库连接"
	@echo "  db-verify     - 验证数据库表结构"
	@echo "  admin-check   - 检查和修复管理员密码"
	@echo "  debug-data    - 检查表中的数据统计"

db-migrate:
	@echo "执行数据库迁移..."
	@go run db/migrate/migrate.go

db-test:
	@echo "测试数据库连接..."
	@go run db/test/test.go

db-verify:
	@echo "验证数据库表结构..."
	@go run db/verify/verify.go

admin-check:
	@echo "检查管理员账户..."
	@go run admin/check_password.go

debug-data:
	@echo "检查表数据..."
	@go run debug/check_data.go