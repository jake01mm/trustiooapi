# Trusioo Card Detection API Postman 集合

本文档介绍如何使用 Trusioo 卡片检测 API 的 Postman 集合进行接口测试。

## 文件结构

```
Postman/auth/
├── Trusioo_CardDetection.postman_collection.json  # 主集合文件
├── Trusioo_CardDetection.postman_environment.json # 环境变量文件
└── CardDetection_README.md                        # 使用说明文档
```

## 快速开始

### 1. 导入集合和环境

1. 打开 Postman
2. 点击 "Import" 按钮
3. 导入 `Trusioo_CardDetection.postman_collection.json`
4. 导入 `Trusioo_CardDetection.postman_environment.json`
5. 在右上角选择 "Trusioo Card Detection Environment" 环境

### 2. 配置环境变量

在环境变量中设置以下必要参数：

- `base_url`: API 基础地址 (默认: http://localhost:8080)
- `access_token`: 用户访问令牌 (需要先通过用户认证获取)

### 3. 推荐测试执行流程

1. **获取支持的地区列表** - 获取可用的检测地区
2. **获取服务状态** - 确认服务正常运行
3. **提交卡片检测** - 提交卡片进行检测
4. **查询检测结果** - 获取检测结果
5. **获取用户检测历史** - 查看历史记录
6. **获取检测记录详情** - 查看具体记录详情
7. **获取用户统计信息** - 查看统计数据
8. **获取用户汇总信息** - 查看汇总信息

## API 接口详细说明

### 公共接口 (无需认证)

#### 1. 获取支持的地区列表
- **方法**: GET
- **路径**: `/api/v1/card-detection/regions?productMark=iTunes`
- **描述**: 获取卡片检测支持的地区列表
- **必需参数**: `productMark` (查询参数) - 产品标识，支持值：iTunes, Amazon, Razer, Xbox
- **响应**: 地区信息数组，包含 ID、名称、代码等

#### 2. 获取服务状态
- **方法**: GET
- **路径**: `/api/v1/card-detection/status`
- **描述**: 获取卡片检测服务的运行状态
- **响应**: 服务状态、版本、运行时间等信息

### 卡片检测接口 (需要认证)

#### 3. 提交卡片检测
- **方法**: POST
- **路径**: `/api/v1/card-detection/check`
- **认证**: Bearer Token
- **请求体**:
  ```json
  {
    "cards": [
      {
        "card_no": "1234567890123456",
        "pin_code": "1234"
      }
    ],
    "product_mark": "iTunes",
    "region_id": 1,
    "region_name": "美国",
    "auto_type": true
  }
  ```
- **描述**: 提交卡片检测请求
- **响应**: 返回请求ID和状态

#### 4. 查询检测结果
- **方法**: POST
- **路径**: `/api/v1/card-detection/result`
- **认证**: Bearer Token
- **请求体**:
  ```json
  {
    "product_mark": "iTunes",
    "card_no": "1234567890123456",
    "pin_code": "1234"
  }
  ```
- **描述**: 查询卡片检测结果
- **响应**: 返回加密的检测结果

### 历史记录与统计接口 (需要认证)

#### 5. 获取用户检测历史
- **方法**: GET
- **路径**: `/api/v1/card-detection/history`
- **认证**: Bearer Token
- **查询参数**:
  - `page`: 页码 (必需)
  - `page_size`: 每页数量 (必需, 1-100)
  - `status`: 状态过滤 (可选)
  - `product_mark`: 产品类型过滤 (可选)
- **描述**: 获取用户的卡片检测历史记录
- **响应**: 分页的历史记录列表

#### 6. 获取检测记录详情
- **方法**: GET
- **路径**: `/api/v1/card-detection/records/{record_id}`
- **认证**: Bearer Token
- **路径参数**: `record_id` - 记录ID
- **描述**: 获取单个检测记录的详细信息
- **响应**: 详细的记录信息

#### 7. 获取用户统计信息
- **方法**: GET
- **路径**: `/api/v1/card-detection/stats`
- **认证**: Bearer Token
- **描述**: 获取用户的检测统计信息
- **响应**: 包含汇总、产品统计、月度统计等

#### 8. 获取用户汇总信息
- **方法**: GET
- **路径**: `/api/v1/card-detection/summary`
- **认证**: Bearer Token
- **描述**: 获取用户的检测汇总信息
- **响应**: 总检测次数、成功次数、成功率等

## 自动化功能

### 环境变量自动管理

集合包含自动化脚本，可以：

1. **自动保存地区ID**: 从地区列表接口自动保存第一个地区ID到环境变量 `region_id`
2. **自动保存产品标识**: 从地区列表请求中保存产品标识到 `product_mark`
3. **自动保存请求ID**: 提交检测后自动保存请求ID
4. **自动保存卡号和产品标识**: 从检测请求中提取并保存
5. **自动保存记录ID**: 从历史记录中自动保存第一条记录ID
6. **自动保存加密结果**: 从检测结果中保存加密数据

### 测试脚本功能

每个接口都包含以下自动化测试：

1. **状态码验证**: 确保返回 200 状态码
2. **响应结构验证**: 验证响应包含必要字段
3. **数据有效性验证**: 验证返回数据的格式和内容
4. **性能测试**: 检查响应时间是否在合理范围内

## 使用技巧

### 1. 认证令牌获取

在使用需要认证的接口前，需要先通过用户认证接口获取 `access_token`：

1. 使用用户认证集合登录
2. 复制获取的 `access_token`
3. 在卡片检测环境中设置 `access_token` 变量

### 2. 批量测试

可以使用 Postman 的 Collection Runner 功能进行批量测试：

1. 点击集合右侧的 "..." 菜单
2. 选择 "Run collection"
3. 选择要运行的请求
4. 设置迭代次数和延迟
5. 点击 "Run" 开始测试

### 3. 数据驱动测试

可以准备 CSV 文件包含不同的测试数据：

```csv
card_no,pin_code,product_mark,region_id
1234567890123456,1234,iTunes,1
2345678901234567,2345,Sephora,2
3456789012345678,3456,Razer,3
```

然后在 Collection Runner 中上传 CSV 文件进行数据驱动测试。

## 故障排除

### 常见问题

1. **401 Unauthorized**
   - 检查 `access_token` 是否正确设置
   - 确认令牌是否已过期
   - 重新获取用户认证令牌

2. **404 Not Found**
   - 检查 `base_url` 是否正确
   - 确认 API 路径是否正确
   - 检查服务是否正在运行

3. **400 Bad Request**
   - 对于 `/regions` 接口：确保提供了 `productMark` 查询参数
   - 对于POST请求：检查请求体格式是否正确
   - 确认必需参数是否都已提供
   - 验证参数值是否有效
   - 避免在GET请求中添加 `Content-Type: application/json` 头

4. **500 Internal Server Error**
   - 检查服务器日志
   - 确认数据库连接是否正常
   - 联系开发团队

### 调试技巧

1. **查看请求详情**: 在 Postman Console 中查看完整的请求和响应
2. **使用变量**: 利用环境变量和全局变量管理测试数据
3. **添加日志**: 在测试脚本中添加 `console.log()` 输出调试信息
4. **分步测试**: 逐个测试接口，确保每步都成功

## 扩展功能

### 1. 监控和报告

可以集成 Postman Monitor 功能：

1. 设置定期运行的监控
2. 配置失败通知
3. 生成测试报告

### 2. CI/CD 集成

可以使用 Newman (Postman CLI) 在 CI/CD 流水线中运行测试：

```bash
newman run Trusioo_CardDetection.postman_collection.json \
  -e Trusioo_CardDetection.postman_environment.json \
  --reporters cli,json,html
```

### 3. 性能测试

可以配置更严格的性能测试：

```javascript
pm.test('Response time is acceptable', function () {
    pm.expect(pm.response.responseTime).to.be.below(500);
});
```

## 版本历史

- **v1.0.0** (2024-01-20)
  - 初始版本
  - 包含所有卡片检测 API 接口
  - 支持自动化测试和环境变量管理
  - 提供完整的使用文档

## 支持

如有问题或建议，请联系开发团队或在项目仓库中提交 Issue。