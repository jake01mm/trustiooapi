.PHONY: help dev build run test clean install-air migrate-up migrate-down migrate-status backup kill

# 默认目标
help:
	@echo "Available commands:"
	@echo "  dev          - Start development server with hot reload (Air)"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-air  - Install Air hot reload tool"
	@echo "  init-db      - Initialize the database"
	@echo "  backup       - Create database backup"
	@echo "  kill         - Kill process listening on port 8080"

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


# 数据库初始化
init-db:
	@echo "Initializing database..."
	@docker exec -i trusioo-postgres psql -U neondb_owner -d neondb < scripts/init_db.sql

# 数据库备份
backup:
	@echo "Creating database backup..."
	@go run tools/db/backup/backup.go

# Kill process on port 8080 (macOS)
kill:
	@echo "Killing process on port 8080..."
	@PIDS=$$(lsof -ti tcp:8080 -sTCP:LISTEN || true); \
	if [ -n "$$PIDS" ]; then \
		echo "Found PID(s): $$PIDS"; \
		kill -15 $$PIDS || true; \
		sleep 1; \
		STILL=$$(lsof -ti tcp:8080 -sTCP:LISTEN || true); \
		if [ -n "$$STILL" ]; then \
			echo "Force killing: $$STILL"; \
			kill -9 $$STILL || true; \
		fi; \
	else \
		echo "No process is listening on port 8080"; \
	fi; \
	FINAL=$$(lsof -ti tcp:8080 -sTCP:LISTEN || true); \
	if [ -z "$$FINAL" ]; then \
		echo "Port 8080 is now free."; \
	else \
		echo "Port 8080 still has listeners: $$FINAL"; \
	fi