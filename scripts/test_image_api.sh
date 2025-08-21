#!/bin/bash

# Trusioo API 图片功能测试脚本

set -e

# 配置
API_BASE_URL="http://localhost:8080"
TEST_IMAGE_PATH="test_images/sample.jpg"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 创建测试图片
create_test_image() {
    log_info "创建测试图片..."
    mkdir -p test_images
    
    if command -v convert &> /dev/null; then
        # 使用ImageMagick创建测试图片
        convert -size 800x600 xc:lightblue \
            -pointsize 60 -fill black -gravity center \
            -annotate +0+0 "Trusioo Test Image" \
            "$TEST_IMAGE_PATH"
        log_success "测试图片已创建: $TEST_IMAGE_PATH"
    else
        # 创建一个简单的测试文件
        echo "This is a test image file for Trusioo API" > "$TEST_IMAGE_PATH"
        log_warning "ImageMagick未安装，创建了文本测试文件"
    fi
}

# 测试服务器连接
test_server_connection() {
    log_info "测试服务器连接..."
    
    if curl -s "$API_BASE_URL/health" > /dev/null; then
        log_success "服务器连接正常"
        return 0
    else
        log_error "服务器连接失败，请确保API服务器正在运行"
        log_error "启动服务器: ./scripts/start_dev.sh"
        return 1
    fi
}

# 测试健康检查端点
test_health_endpoints() {
    log_info "测试健康检查端点..."
    
    # 基本健康检查
    response=$(curl -s "$API_BASE_URL/health")
    if echo "$response" | grep -q "ok\|healthy\|success"; then
        log_success "基本健康检查通过"
    else
        log_warning "基本健康检查响应: $response"
    fi
    
    # 详细健康检查
    response=$(curl -s "$API_BASE_URL/api/v1/health/detailed")
    if [ $? -eq 0 ]; then
        log_success "详细健康检查通过"
    else
        log_warning "详细健康检查失败"
    fi
    
    # 数据库健康检查
    response=$(curl -s "$API_BASE_URL/api/v1/health/database")
    if [ $? -eq 0 ]; then
        log_success "数据库健康检查通过"
    else
        log_warning "数据库健康检查失败，请检查数据库连接"
    fi
}

# 测试图片上传（公开）
test_public_image_upload() {
    log_info "测试公开图片上传..."
    
    if [ ! -f "$TEST_IMAGE_PATH" ]; then
        create_test_image
    fi
    
    response=$(curl -s -X POST "$API_BASE_URL/api/v1/images/upload" \
        -F "file=@$TEST_IMAGE_PATH" \
        -F "is_public=true" \
        -F "folder=test" \
        -F "file_name=public_test_image.jpg")
    
    if echo "$response" | grep -q '"id"'; then
        log_success "公开图片上传成功"
        # 提取图片ID
        IMAGE_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        IMAGE_URL=$(echo "$response" | grep -o '"url":"[^"]*"' | cut -d'"' -f4)
        log_info "图片ID: $IMAGE_ID"
        log_info "图片URL: $IMAGE_URL"
        return 0
    else
        log_error "公开图片上传失败"
        log_error "响应: $response"
        return 1
    fi
}

# 测试图片上传（私有）
test_private_image_upload() {
    log_info "测试私有图片上传..."
    
    response=$(curl -s -X POST "$API_BASE_URL/api/v1/images/upload" \
        -F "file=@$TEST_IMAGE_PATH" \
        -F "is_public=false" \
        -F "folder=private")
    
    if echo "$response" | grep -q '"id"'; then
        log_success "私有图片上传成功"
        PRIVATE_IMAGE_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        log_info "私有图片ID: $PRIVATE_IMAGE_ID"
        return 0
    else
        log_error "私有图片上传失败"
        log_error "响应: $response"
        return 1
    fi
}

# 测试图片列表
test_image_list() {
    log_info "测试图片列表..."
    
    response=$(curl -s "$API_BASE_URL/api/v1/images/?page=1&page_size=10")
    
    if echo "$response" | grep -q '"images"'; then
        image_count=$(echo "$response" | grep -o '"total":[0-9]*' | cut -d':' -f2)
        log_success "图片列表获取成功，共 $image_count 张图片"
        return 0
    else
        log_error "图片列表获取失败"
        log_error "响应: $response"
        return 1
    fi
}

# 测试获取单张图片
test_get_image() {
    if [ -z "$IMAGE_ID" ]; then
        log_warning "跳过单张图片测试（无图片ID）"
        return 0
    fi
    
    log_info "测试获取单张图片 (ID: $IMAGE_ID)..."
    
    response=$(curl -s "$API_BASE_URL/api/v1/images/$IMAGE_ID")
    
    if echo "$response" | grep -q '"id"'; then
        log_success "单张图片获取成功"
        return 0
    else
        log_error "单张图片获取失败"
        log_error "响应: $response"
        return 1
    fi
}

# 测试刷新URL
test_refresh_url() {
    if [ -z "$PRIVATE_IMAGE_ID" ]; then
        log_warning "跳过URL刷新测试（无私有图片ID）"
        return 0
    fi
    
    log_info "测试刷新私有图片URL (ID: $PRIVATE_IMAGE_ID)..."
    
    response=$(curl -s -X PUT "$API_BASE_URL/api/v1/images/$PRIVATE_IMAGE_ID/refresh")
    
    if echo "$response" | grep -q '"url"'; then
        log_success "私有图片URL刷新成功"
        return 0
    else
        log_error "私有图片URL刷新失败"
        log_error "响应: $response"
        return 1
    fi
}

# 测试错误处理
test_error_handling() {
    log_info "测试错误处理..."
    
    # 测试无文件上传
    response=$(curl -s -X POST "$API_BASE_URL/api/v1/images/upload")
    if echo "$response" | grep -q "error\|Error"; then
        log_success "无文件上传错误处理正确"
    else
        log_warning "无文件上传错误处理可能有问题"
    fi
    
    # 测试不存在的图片ID
    response=$(curl -s "$API_BASE_URL/api/v1/images/99999")
    if echo "$response" | grep -q "error\|Error\|not found"; then
        log_success "不存在图片错误处理正确"
    else
        log_warning "不存在图片错误处理可能有问题"
    fi
}

# 性能测试
test_performance() {
    log_info "测试API性能..."
    
    # 测试健康检查响应时间
    start_time=$(date +%s%N)
    curl -s "$API_BASE_URL/health" > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    if [ $duration -lt 100 ]; then
        log_success "健康检查响应时间: ${duration}ms (优秀)"
    elif [ $duration -lt 500 ]; then
        log_success "健康检查响应时间: ${duration}ms (良好)"
    else
        log_warning "健康检查响应时间: ${duration}ms (需要优化)"
    fi
}

# 清理测试数据
cleanup() {
    log_info "清理测试数据..."
    
    # 删除上传的测试图片
    if [ ! -z "$IMAGE_ID" ]; then
        curl -s -X DELETE "$API_BASE_URL/api/v1/images/$IMAGE_ID" > /dev/null
        log_success "已删除公开测试图片 (ID: $IMAGE_ID)"
    fi
    
    if [ ! -z "$PRIVATE_IMAGE_ID" ]; then
        curl -s -X DELETE "$API_BASE_URL/api/v1/images/$PRIVATE_IMAGE_ID" > /dev/null
        log_success "已删除私有测试图片 (ID: $PRIVATE_IMAGE_ID)"
    fi
    
    # 删除测试文件
    if [ -f "$TEST_IMAGE_PATH" ]; then
        rm -f "$TEST_IMAGE_PATH"
        log_success "已删除本地测试图片"
    fi
    
    if [ -d "test_images" ]; then
        rmdir test_images 2>/dev/null || true
    fi
}

# 主测试函数
run_tests() {
    echo "🧪 Trusioo API 图片功能测试开始"
    echo "================================"
    
    # 基础连接测试
    if ! test_server_connection; then
        exit 1
    fi
    
    # 健康检查测试
    test_health_endpoints
    
    # 图片功能测试
    test_public_image_upload
    test_private_image_upload
    test_image_list
    test_get_image
    test_refresh_url
    
    # 错误处理测试
    test_error_handling
    
    # 性能测试
    test_performance
    
    echo ""
    echo "================================"
    log_success "所有测试完成！"
    
    # 显示使用说明
    echo ""
    echo "📚 API使用说明："
    echo "----------------"
    echo "上传图片："
    echo "curl -X POST $API_BASE_URL/api/v1/images/upload \\"
    echo "  -F \"file=@your-image.jpg\" \\"
    echo "  -F \"is_public=true\" \\"
    echo "  -F \"folder=uploads\""
    echo ""
    echo "获取图片列表："
    echo "curl $API_BASE_URL/api/v1/images/?page=1&page_size=20"
    echo ""
    echo "获取单张图片："
    echo "curl $API_BASE_URL/api/v1/images/{id}"
    echo ""
}

# 捕获退出信号，执行清理
trap cleanup EXIT

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-cleanup)
            trap - EXIT
            shift
            ;;
        --base-url)
            API_BASE_URL="$2"
            shift 2
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo "选项:"
            echo "  --no-cleanup     不清理测试数据"
            echo "  --base-url URL   指定API基础URL (默认: http://localhost:8080)"
            echo "  --help          显示此帮助信息"
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            exit 1
            ;;
    esac
done

# 运行测试
run_tests