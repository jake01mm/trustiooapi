# Trusioo API Postman Collections

本目录包含完整的Postman集合，用于测试Trusioo API的各个模块。

## 文件结构

```
Postman/
├── README.md                                    # 本文档
├── Environment_Trusioo_API.postman_environment.json   # 环境变量配置
├── User_Auth_Module.postman_collection.json    # 用户认证模块集合
├── Admin_Auth_Module.postman_collection.json   # 管理员认证模块集合
└── Card_Detection_Module.postman_collection.json # 卡片检测模块集合
```

## 快速开始

### 1. 导入到Postman

1. 打开Postman应用
2. 点击左上角的"Import"按钮
3. 选择"File"标签
4. 导入以下文件：
   - `Environment_Trusioo_API.postman_environment.json` (环境变量)
   - `User_Auth_Module.postman_collection.json` (用户认证集合)
   - `Admin_Auth_Module.postman_collection.json` (管理员认证集合)
   - `Card_Detection_Module.postman_collection.json` (卡片检测集合)

### 2. 设置环境

1. 在Postman右上角选择"Trusioo API Environment"环境
2. 确保以下环境变量已正确设置：
   - `base_url`: http://localhost:8080 (或你的服务器地址)
   - `api_version`: v1
   - `user_password`: securepassword123

### 3. 启动服务器

确保Trusioo API服务器正在运行：
```bash
cd /Users/laitsim/trusioo_api
go run cmd/main.go
```

## User Auth Module 使用指南

### 测试流程顺序

按照以下顺序执行请求以获得最佳测试体验：

1. **User Registration** - 注册新用户
2. **User Login (Step 1)** - 发送登录验证码
3. **User Login (Step 2)** - 验证验证码并获取Token
4. **Get User Profile** - 获取用户资料
5. **Refresh Token** - 刷新访问令牌
6. **Forgot Password** - 发送密码重置验证码
7. **Reset Password** - 重置密码
8. **Login with New Password** - 用新密码登录

### 重要注意事项

#### ⚠️ 验证码获取
系统通过控制台输出验证码（开发环境）。执行需要验证码的请求后：

1. 检查服务器控制台日志
2. 查找类似以下格式的输出：
   ```
   === 验证码发送 ===
   邮箱: user@example.com
   验证码: 123456
   类型: user_login
   ==================
   ```
3. 手动设置Postman环境变量：`verification_code = 123456`

#### 🔄 自动化功能
- **唯一邮箱生成**: 每次注册会自动生成唯一邮箱地址
- **Token管理**: 登录成功后自动保存访问令牌和刷新令牌
- **环境变量更新**: 测试脚本会自动更新相关环境变量

#### 📝 测试验证
每个请求都包含详细的测试脚本，验证：
- HTTP状态码
- 响应结构和数据类型
- 业务逻辑正确性
- 环境变量自动设置

### 环境变量说明

| 变量名 | 描述 | 自动设置 |
|--------|------|----------|
| `base_url` | API基础URL | ❌ 手动 |
| `api_version` | API版本 | ❌ 手动 |
| `user_password` | 测试用户密码 | ❌ 手动 |
| `unique_user_email` | 当前注册的唯一邮箱 | ✅ 自动 |
| `registered_user_email` | 已注册用户邮箱 | ✅ 自动 |
| `user_access_token` | 用户访问令牌 | ✅ 自动 |
| `user_refresh_token` | 用户刷新令牌 | ✅ 自动 |
| `user_id` | 用户ID | ✅ 自动 |
| `verification_code` | 验证码 | ❌ 手动 |
| `new_password` | 重置后的新密码 | ❌ 手动 |

## Admin Auth Module 使用指南

### 测试流程顺序

按照以下顺序执行请求以获得最佳测试体验：

1. **Admin Login (Step 1)** - 发送管理员登录验证码
2. **Admin Login (Step 2)** - 验证验证码并获取Token
3. **Get Admin Profile** - 获取管理员资料
4. **Admin Refresh Token** - 刷新管理员访问令牌
5. **Get User Statistics** - 获取用户统计信息
6. **Get User List** - 获取用户列表
7. **Get User Detail** - 获取特定用户详情
8. **Admin Forgot Password** - 发送密码重置验证码
9. **Admin Reset Password** - 重置管理员密码

### 重要注意事项

#### 🔑 默认管理员账户
- **邮箱**: admin@trusioo.com
- **密码**: admin123 (来自数据库迁移的默认账户)
- **角色**: super_admin

#### ⚠️ 管理员验证码获取
与用户登录类似，管理员验证码也通过控制台输出。执行需要验证码的请求后：
1. 检查服务器控制台日志
2. 查找管理员验证码输出
3. 手动设置Postman环境变量：`admin_verification_code = 123456`

#### 🛡️ 权限控制
- 所有用户管理端点需要管理员权限
- 使用 `admin_access_token` 进行身份验证
- 管理员和用户的Token是独立的

## Card Detection Module 使用指南

### 测试流程顺序

按照以下顺序执行请求以获得最佳测试体验：

1. **Get CD Products** - 获取可用产品列表
2. **Get CD Regions** - 获取特定产品的区域列表
3. **Check Card** - 提交卡片进行检测
4. **Check Card Result** - 获取检测结果
5. **Get User History** - 查看检测历史
6. **Get Record Detail** - 获取特定记录详情
7. **Get User Stats** - 获取详细统计信息
8. **Get User Summary** - 获取快速汇总

### 重要注意事项

#### 🎫 支持的产品类型
- **iTunes**: 苹果礼品卡，支持多个国家/地区
- **Amazon**: 亚马逊礼品卡，支持美亚/加亚、欧盟区
- **Razer**: 雷蛇礼品卡，支持多个地区
- **Xbox**: XBOX礼品卡，支持多个国家
- **Sephora**: 丝芙兰礼品卡
- **Nike**: NIKE礼品卡

#### 🌍 区域要求
- 部分产品需要指定区域（如iTunes、Amazon、Razer）
- 首先调用 Get CD Products 了解产品配置
- 然后调用 Get CD Regions 获取可用区域

#### 🔐 认证要求
- 所有卡片检测端点都需要用户或管理员认证
- 可以使用 `user_access_token` 或 `admin_access_token`
- 测试脚本会自动选择可用的Token

#### 🔄 异步检测流程
1. **Check Card**: 提交卡片检测请求，获得 request_id
2. **Check Card Result**: 使用产品标识和卡号查询结果
3. **历史记录**: 所有检测都会记录在用户历史中

#### 📊 测试数据配置
环境变量中预设了测试数据：
- `test_card_number`: 1234567890123456
- `test_product_mark`: iTunes
- `test_region_id`: 2 (美国)
- `test_pin_code`: 1234

## API端点覆盖

### User Auth Module ✅

| 端点 | 方法 | 描述 | 状态 |
|------|------|------|------|
| `/api/v1/auth/register` | POST | 用户注册 | ✅ |
| `/api/v1/auth/login` | POST | 登录步骤1 - 发送验证码 | ✅ |
| `/api/v1/auth/login/verify` | POST | 登录步骤2 - 验证验证码 | ✅ |
| `/api/v1/auth/profile` | GET | 获取用户资料 | ✅ |
| `/api/v1/auth/refresh` | POST | 刷新访问令牌 | ✅ |
| `/api/v1/auth/forgot-password` | POST | 忘记密码 | ✅ |
| `/api/v1/auth/reset-password` | POST | 重置密码 | ✅ |

### Admin Auth Module ✅

| 端点 | 方法 | 描述 | 状态 |
|------|------|------|------|
| `/api/v1/admin/auth/login` | POST | 管理员登录步骤1 - 发送验证码 | ✅ |
| `/api/v1/admin/auth/login/verify` | POST | 管理员登录步骤2 - 验证验证码 | ✅ |
| `/api/v1/admin/profile` | GET | 获取管理员资料 | ✅ |
| `/api/v1/admin/auth/refresh` | POST | 刷新管理员访问令牌 | ✅ |
| `/api/v1/admin/auth/forgot-password` | POST | 管理员忘记密码 | ✅ |
| `/api/v1/admin/auth/reset-password` | POST | 管理员重置密码 | ✅ |
| `/api/v1/admin/users/stats` | GET | 获取用户统计信息 | ✅ |
| `/api/v1/admin/users` | GET | 获取用户列表 | ✅ |
| `/api/v1/admin/users/:id` | GET | 获取用户详情 | ✅ |

### Card Detection Module ✅

| 端点 | 方法 | 描述 | 状态 |
|------|------|------|------|
| `/api/v1/card-detection/cd_products` | GET | 获取CD产品列表 | ✅ |
| `/api/v1/card-detection/cd_regions` | GET | 获取CD区域列表 | ✅ |
| `/api/v1/card-detection/check` | POST | 提交卡片检测 | ✅ |
| `/api/v1/card-detection/result` | POST | 获取检测结果 | ✅ |
| `/api/v1/card-detection/history` | GET | 获取检测历史 | ✅ |
| `/api/v1/card-detection/records/:id` | GET | 获取检测记录详情 | ✅ |
| `/api/v1/card-detection/stats` | GET | 获取用户检测统计 | ✅ |
| `/api/v1/card-detection/summary` | GET | 获取用户检测汇总 | ✅ |

## 故障排除

### 常见问题

1. **401 Unauthorized**
   - **用户端点**: 检查访问令牌是否已设置：`{{user_access_token}}`
   - **管理员端点**: 检查管理员令牌是否已设置：`{{admin_access_token}}`
   - **卡片检测**: 确保已登录用户或管理员，令牌会自动选择
   - 确保令牌没有过期，必要时使用刷新令牌

2. **验证码无效**
   - 确保从服务器日志中复制了正确的验证码
   - 检查验证码是否已过期（600秒有效期）
   - 确保验证码类型匹配：
     - 用户登录: `user_login`
     - 用户忘记密码: `forgot_password`
     - 管理员登录: `admin_login`
     - 管理员忘记密码: `admin_forgot_password`

3. **邮箱已存在**
   - 注册请求会自动生成唯一邮箱
   - 如果仍然失败，清除`unique_user_email`环境变量重试

4. **服务器连接失败**
   - 确保API服务器在`http://localhost:8080`运行
   - 检查`base_url`环境变量设置
   - 验证防火墙和网络连接

5. **管理员登录失败**
   - 确认使用正确的管理员邮箱：`admin@trusioo.com`
   - 确认使用正确的管理员密码：`admin123`
   - 检查数据库中是否存在默认管理员账户

6. **卡片检测失败**
   - 确保产品类型正确（iTunes, amazon, Razer等）
   - 检查是否需要指定区域（调用Get CD Products查看）
   - 验证卡号格式符合产品要求
   - 确保使用正确的PIN码（如果产品需要）

7. **获取区域列表为空**
   - 确保指定了正确的product_mark参数
   - 检查产品是否支持区域选择
   - 验证产品状态是否为active

8. **检测记录不存在**
   - 确保使用正确的record_id
   - 检查用户是否有权限访问该记录
   - 确认记录属于当前登录用户

### 调试技巧

1. **查看控制台输出**
   - Postman控制台 (View > Show Postman Console)
   - 服务器控制台日志

2. **检查环境变量**
   - 点击环境名称旁的眼睛图标
   - 验证所有必需变量都已正确设置

3. **逐步测试**
   - 按推荐顺序执行请求
   - 每步验证响应和环境变量更新

## 贡献

如果发现问题或有改进建议：
1. 检查服务器日志获取详细错误信息
2. 验证环境变量设置
3. 确保按正确顺序执行请求
4. 报告问题时请包含具体的错误信息和重现步骤