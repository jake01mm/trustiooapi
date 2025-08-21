.PHONY: help dev build run test clean install-air migrate-up migrate-down migrate-status backup

# 默认目标
help:
	@echo "Available commands:"
	@echo "  dev          - Start development server with hot reload (Air)"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-air  - Install Air hot reload tool"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  migrate-status - Check migration status"
	@echo "  backup       - Create database backup"

# 开发模式 - 使用Air热重载
dev:
	@echo "Starting development server with hot reload..."
	@if ! command -v air > /dev/null; then \
		echo "Air not found. Installing..."; \
		make install-air; \
	fi
	@air

# 构建应用
build:
	@echo "Building application..."
	@go build -o build/trusioo-api ./cmd/main.go

# 运行应用
run:
	@echo "Running application..."
	@go run cmd/main.go

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 清理构建文件
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf build/
	@rm -rf tmp/
	@rm -f build-errors.log

# 安装Air热重载工具
install-air:
	@echo "Installing Air..."
	@go install github.com/cosmtrek/air@latest

# 数据库迁移 - 向上
migrate-up:
	@echo "Running database migrations up..."
	@go run tools/db/migrate/migrate.go up

# 数据库迁移 - 向下
migrate-down:
	@echo "Running database migrations down..."
	@go run tools/db/migrate/migrate.go down

# 检查迁移状态
migrate-status:
	@echo "Checking migration status..."
	@go run tools/db/migrate/migrate.go status

# 数据库备份
backup:
	@echo "Creating database backup..."
	@go run tools/db/backup/backup.go