# 卡片检测模块重构完成

## 🎯 重构目标

重新设计并实现卡片检测模块，解决之前版本存在的架构问题，实现真正的异步处理机制。

## 🏗️ 新架构设计

### 目录结构
```
internal/carddetection/
├── admin/               # 管理员接口层
│   ├── handler.go      # 管理员处理器
│   └── routes.go       # 管理员路由
├── user/               # 用户接口层  
│   ├── handler.go      # 用户处理器
│   └── routes.go       # 用户路由
├── shared/             # 共享组件层
│   ├── entities.go     # 数据模型定义
│   ├── dto.go          # 数据传输对象
│   ├── errors.go       # 错误定义
│   ├── repository.go   # 数据访问层
│   ├── service.go      # 业务逻辑层
│   └── processor.go    # 异步任务处理器
└── module.go           # 模块入口
```

### 核心组件

#### 1. 数据模型 (Entities)
- **CardDetectionTask**: 检测任务实体
- **CardDetectionRecord**: 卡片检测记录实体
- **CardDetectionCache**: 结果缓存实体

#### 2. 业务流程
1. **提交阶段**: 用户提交检测任务，立即返回task_id
2. **处理阶段**: 后台异步处理器调用第三方API
3. **查询阶段**: 用户通过task_id查询处理状态和结果

#### 3. 异步处理机制
- **TaskProcessor**: 多线程任务处理器
- **工作线程池**: 可配置的并发处理能力
- **智能重试**: 自动处理失败的请求
- **缓存优化**: 避免重复查询相同卡片

## 📊 数据库设计

### 新增数据表

#### card_detection_tasks (检测任务表)
```sql
CREATE TABLE card_detection_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(36) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    user_type VARCHAR(20) NOT NULL, -- 'user' | 'admin'
    product_mark VARCHAR(20) NOT NULL,
    region_id INT,
    region_name VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority VARCHAR(10) DEFAULT 'normal',
    total_cards INT NOT NULL DEFAULT 0,
    completed_cards INT NOT NULL DEFAULT 0,
    failed_cards INT NOT NULL DEFAULT 0,
    -- 时间戳字段
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### card_detection_records (检测记录表)
```sql
CREATE TABLE card_detection_records (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    card_no VARCHAR(100) NOT NULL,
    pin_code VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- 第三方API结果
    card_status INT,
    card_status_name VARCHAR(50),
    message TEXT,
    balance VARCHAR(50),
    check_time TIMESTAMP,
    region_name VARCHAR(50),
    region_id INT,
    
    -- 性能指标
    response_time INT,
    retry_count INT DEFAULT 0,
    last_error TEXT,
    
    -- 时间戳
    submitted_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### card_detection_cache (结果缓存表)
```sql
CREATE TABLE card_detection_cache (
    id SERIAL PRIMARY KEY,
    card_no VARCHAR(100) NOT NULL,
    product_mark VARCHAR(20) NOT NULL,
    pin_code_hash VARCHAR(64),
    
    -- 缓存的结果
    card_status INT NOT NULL,
    card_status_name VARCHAR(50),
    message TEXT,
    balance VARCHAR(50),
    check_time TIMESTAMP,
    region_name VARCHAR(50),
    region_id INT,
    
    -- 缓存元信息
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    hit_count INT DEFAULT 0,
    last_hit_at TIMESTAMP
);
```

## 🚀 API接口设计

### 用户接口 (`/api/v1/carddetection/user/`)
```
POST   /tasks              提交检测任务
GET    /tasks/:taskId       获取任务状态
GET    /tasks/:taskId/results 获取任务结果
GET    /tasks              获取任务历史
POST   /query              直接查询卡片
GET    /products           获取产品和区域信息
GET    /stats              获取用户统计信息
```

### 管理员接口 (`/api/v1/carddetection/admin/`)
```
# 基础功能 (与用户接口对等)
POST   /tasks              提交检测任务
GET    /tasks/:taskId       获取任务状态
GET    /tasks/:taskId/results 获取任务结果
GET    /tasks              获取任务历史
POST   /query              直接查询卡片
GET    /products           获取产品和区域信息
GET    /stats              获取管理员统计信息

# 管理员专有功能
GET    /system/stats       获取系统统计信息
GET    /system/history     获取所有用户检测历史
GET    /system/users/:userId 获取指定用户详情
```

## 💡 核心特性

### 1. 真正异步处理
- ✅ 提交任务立即返回，无需等待
- ✅ 后台多线程并行处理
- ✅ 实时进度跟踪
- ✅ 任务状态透明可见

### 2. 智能缓存机制
- ✅ 按结果类型设置不同缓存期限
- ✅ 自动清理过期缓存
- ✅ 缓存命中统计
- ✅ 避免重复查询费用

### 3. 容错与重试
- ✅ 自动重试失败的请求
- ✅ 可配置的重试次数和策略
- ✅ 详细的错误信息记录
- ✅ 优雅的异常处理

### 4. 性能优化
- ✅ 批量处理减少API调用
- ✅ 工作线程池可配置
- ✅ 数据库连接池优化
- ✅ 索引优化查询性能

### 5. 监控与统计
- ✅ 任务执行统计
- ✅ 成功率分析
- ✅ 响应时间监控
- ✅ 缓存命中率统计

## 🔧 配置说明

### 环境变量
```bash
# 卡片检测服务配置
CARD_DETECTION_ENABLED=true
CARD_DETECTION_HOST=https://ckxiang.com
CARD_DETECTION_APP_ID=your_app_id
CARD_DETECTION_APP_SECRET=your_app_secret
CARD_DETECTION_TIMEOUT=30
```

### 任务处理器配置
```go
// 在router初始化时配置
processor := cardDetectionModule.GetProcessor()
processor.SetWorkerCount(10)        // 设置工作线程数
processor.SetRetryLimit(5)          // 设置重试次数
processor.SetBatchSize(100)         // 设置批处理大小
processor.SetPollInterval(5*time.Second) // 设置轮询间隔
```

## 📈 使用示例

### 提交检测任务
```bash
curl -X POST 'http://localhost:8080/api/v1/carddetection/admin/tasks' \
  -H 'Authorization: Bearer TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "cards": [
      {"cardNo": "XPQ2DMF49ZMTZ92Z"},
      {"cardNo": "ANOTHER_CARD_NUMBER"}
    ],
    "productMark": "iTunes",
    "regionId": 1,
    "regionName": "英国",
    "priority": "normal"
  }'

# 返回
{
  "code": 200,
  "message": "success",
  "data": {
    "taskId": "uuid-123-456-789",
    "totalCards": 2,
    "status": "pending",
    "submittedAt": "2025-01-23T12:00:00Z",
    "estimatedCompletion": "2025-01-23T12:05:00Z"
  }
}
```

### 查询任务状态
```bash
curl 'http://localhost:8080/api/v1/carddetection/admin/tasks/uuid-123-456-789' \
  -H 'Authorization: Bearer TOKEN'

# 返回
{
  "code": 200,
  "message": "success",
  "data": {
    "taskId": "uuid-123-456-789",
    "status": "processing",
    "totalCards": 2,
    "completedCards": 1,
    "failedCards": 0,
    "progress": 50.0,
    "submittedAt": "2025-01-23T12:00:00Z",
    "startedAt": "2025-01-23T12:00:30Z",
    "estimatedCompletion": "2025-01-23T12:04:15Z"
  }
}
```

### 获取检测结果
```bash
curl 'http://localhost:8080/api/v1/carddetection/admin/tasks/uuid-123-456-789/results' \
  -H 'Authorization: Bearer TOKEN'

# 返回
{
  "code": 200,
  "message": "success",
  "data": {
    "taskId": "uuid-123-456-789",
    "results": [
      {
        "cardNo": "XPQ2DMF49ZMTZ92Z",
        "status": "completed",
        "cardStatus": 2,
        "cardStatusName": "有效",
        "message": "卡片有效，余额充足",
        "balance": "$50.00",
        "regionName": "英国",
        "checkTime": "2025-01-23T12:01:30Z",
        "responseTime": 1250
      }
    ]
  }
}
```

## 🚀 部署说明

### 1. 运行数据库迁移
```bash
migrate -path migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up
```

### 2. 启动应用
应用启动时会自动：
- 初始化卡片检测模块
- 启动异步任务处理器
- 注册所有API路由

### 3. 监控日志
```bash
tail -f logs/app.log | grep "card detection"
```

## 🔄 从旧版本迁移

### 数据迁移
旧版本的 `card_detection_records` 表已被删除，新的数据结构不兼容。如需保留历史数据，需要：
1. 在迁移前备份旧数据
2. 编写数据转换脚本
3. 将旧数据导入新的表结构

### API兼容性
⚠️ **Breaking Changes**:
- 旧的 `/api/v1/carddetection/check` 接口已移除
- 新的接口路径为 `/api/v1/carddetection/{user|admin}/tasks`
- 请求和响应格式完全不同

### 前端适配
前端需要更新以适配新的API：
- 使用新的任务提交接口
- 实现轮询机制查询任务状态
- 更新结果展示组件

## 🎉 重构效果

### 解决的问题
1. ✅ **同步阻塞**: 现在是真正的异步处理
2. ✅ **错误理解**: 正确理解第三方API工作方式
3. ✅ **状态管理**: 完整的任务生命周期跟踪
4. ✅ **性能问题**: 支持批量处理和并发
5. ✅ **缓存缺失**: 智能缓存避免重复查询

### 带来的优势
1. 🚀 **用户体验**: 提交即返回，实时进度显示
2. 📈 **系统性能**: 支持高并发，资源利用率高
3. 💰 **成本控制**: 缓存机制减少第三方API调用费用
4. 🔧 **易于维护**: 清晰的模块分离，代码结构优良
5. 📊 **业务洞察**: 丰富的统计数据和监控能力

## 🛠️ 后续优化方向

1. **WebSocket实时通知**: 推送任务状态更新
2. **分布式部署**: 支持多实例部署和负载均衡
3. **更多统计功能**: 增加更详细的业务分析
4. **API限流保护**: 防止第三方API调用超限
5. **批量导入**: 支持CSV/Excel文件批量导入卡片

---

✨ **重构完成！新的卡片检测模块已经准备就绪，具备了产品级的稳定性和可扩展性。**