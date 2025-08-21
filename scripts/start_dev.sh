#!/bin/bash

# Trusioo API 开发环境启动脚本

set -e

echo "🚀 Trusioo API 开发环境启动中..."

# 检查Go版本
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.24.0+"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'o' -f2)
echo "✅ Go版本: $GO_VERSION"

# 检查环境变量文件
if [ ! -f ".env" ]; then
    echo "📝 .env文件不存在，从模板复制..."
    cp .env.example .env
    echo "⚠️  请编辑.env文件填入实际配置值"
fi

# 检查数据库连接（可选）
echo "🗄️  检查数据库连接..."
if command -v psql &> /dev/null; then
    DB_HOST=${DB_HOST:-localhost}
    DB_PORT=${DB_PORT:-5432}
    DB_USER=${DB_USER:-postgres}
    DB_NAME=${DB_NAME:-trusioo_db}
    
    if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
        echo "✅ 数据库连接正常"
    else
        echo "⚠️  数据库未连接，某些功能可能不可用"
    fi
else
    echo "⚠️  psql未安装，跳过数据库检查"
fi

# 检查Redis连接（可选）
echo "🔧 检查Redis连接..."
if command -v redis-cli &> /dev/null; then
    REDIS_HOST=${REDIS_HOST:-localhost}
    REDIS_PORT=${REDIS_PORT:-6379}
    
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
        echo "✅ Redis连接正常"
    else
        echo "⚠️  Redis未连接，将使用内存缓存"
    fi
else
    echo "⚠️  Redis未安装，将使用内存缓存"
fi

# 下载依赖
echo "📦 下载Go模块依赖..."
go mod download

# 运行数据库迁移
echo "🗃️  运行数据库迁移..."
if go run tools/db/migrate/main.go -direction=up 2>/dev/null; then
    echo "✅ 数据库迁移完成"
else
    echo "⚠️  数据库迁移失败，请检查数据库配置"
fi

# 构建项目
echo "🔨 构建项目..."
go build -o tmp/trusioo-api cmd/main.go

# 启动服务器
echo ""
echo "🎉 启动Trusioo API服务器..."
echo ""
echo "📋 服务信息:"
echo "   - API地址: http://localhost:${PORT:-8080}"
echo "   - 健康检查: http://localhost:${PORT:-8080}/health"
echo "   - API文档: http://localhost:${PORT:-8080}/api/v1/"
echo ""
echo "🖼️  图片功能测试:"
echo "   - 上传图片: POST http://localhost:${PORT:-8080}/api/v1/images/upload"
echo "   - 图片列表: GET http://localhost:${PORT:-8080}/api/v1/images/"
echo ""
echo "⚠️  注意事项:"
echo "   - 确保已在.env中配置R2存储凭证"
echo "   - 图片上传需要multipart/form-data格式"
echo "   - 支持的图片格式: JPEG, PNG, GIF, WebP"
echo ""
echo "🛑 按 Ctrl+C 停止服务器"
echo ""

# 设置开发环境变量
export GIN_MODE=debug
export ENV=development

# 启动应用
./tmp/trusioo-api