# Trusioo User Auth API Postman 集合

这是 Trusioo 用户认证模块的 Postman API 测试集合，包含用户注册、登录、令牌管理和个人信息获取等功能。

## 文件结构

```
Postman/auth/
├── Trusioo_User_Auth.postman_collection.json  # 用户认证 API 集合
├── Trusioo_User_Auth.postman_environment.json # 环境变量配置
└── User_Auth_README.md                         # 使用说明文档
```

## 快速开始

### 1. 导入集合和环境

1. 打开 Postman
2. 点击 "Import" 按钮
3. 导入 `Trusioo_User_Auth.postman_collection.json`
4. 导入 `Trusioo_User_Auth.postman_environment.json`
5. 在右上角选择 "Trusioo User Auth Environment" 环境

### 2. 配置环境变量

在环境变量中配置以下参数：

- `base_url`: API 服务器地址（默认: http://localhost:8080）
- `user_name`: 测试用户姓名
- `user_email`: 测试用户邮箱
- `user_password`: 测试用户密码
- `user_phone`: 测试用户手机号（可选）

### 3. 推荐测试执行流程

1. **用户注册** → 创建新用户账户
2. **用户登录 - 第一步** → 获取登录验证码
3. **用户登录 - 第二步** → 验证登录码并获取令牌
4. **获取用户个人信息** → 验证认证状态
5. **刷新访问令牌** → 测试令牌刷新功能

## API 接口详细说明

### 用户认证模块

#### 1. 用户注册
- **路径**: `POST /api/v1/auth/register`
- **描述**: 创建新用户账户
- **请求参数**:
  ```json
  {
    "name": "用户姓名",
    "email": "用户邮箱",
    "password": "用户密码",
    "phone": "用户手机号（可选）"
  }
  ```
- **响应示例**:
  ```json
  {
    "data": {
      "user_id": 123,
      "message": "注册成功"
    },
    "message": "用户注册成功"
  }
  ```
- **自动化功能**: 自动保存 `user_id` 到环境变量

#### 2. 用户登录 - 第一步
- **路径**: `POST /api/v1/auth/login`
- **描述**: 发送邮箱和密码，获取登录验证码
- **请求参数**:
  ```json
  {
    "email": "用户邮箱",
    "password": "用户密码"
  }
  ```
- **响应示例**:
  ```json
  {
    "data": {
      "login_code": "123456",
      "expires_in": 300
    },
    "message": "验证码已发送"
  }
  ```
- **自动化功能**: 自动保存 `login_code` 到环境变量

#### 3. 用户登录 - 第二步（验证码验证）
- **路径**: `POST /api/v1/auth/login/verify`
- **描述**: 验证登录验证码，获取访问令牌
- **请求参数**:
  ```json
  {
    "email": "用户邮箱",
    "login_code": "验证码"
  }
  ```
- **响应示例**:
  ```json
  {
    "data": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
      "user": {
        "id": 123,
        "name": "用户姓名",
        "email": "用户邮箱",
        "status": "active"
      }
    },
    "message": "登录成功"
  }
  ```
- **自动化功能**: 自动保存 `access_token`、`refresh_token` 和用户信息到环境变量

### 令牌管理模块

#### 4. 刷新访问令牌
- **路径**: `POST /api/v1/auth/refresh`
- **描述**: 使用刷新令牌获取新的访问令牌
- **请求参数**:
  ```json
  {
    "refresh_token": "刷新令牌"
  }
  ```
- **响应示例**:
  ```json
  {
    "data": {
      "access_token": "新的访问令牌",
      "refresh_token": "新的刷新令牌"
    },
    "message": "令牌刷新成功"
  }
  ```
- **自动化功能**: 自动更新环境变量中的令牌

### 用户信息模块

#### 5. 获取用户个人信息
- **路径**: `GET /api/v1/auth/profile`
- **描述**: 获取当前登录用户的个人信息
- **请求头**: `Authorization: Bearer {access_token}`
- **响应示例**:
  ```json
  {
    "data": {
      "id": 123,
      "name": "用户姓名",
      "email": "用户邮箱",
      "phone": "+1234567890",
      "status": "active",
      "email_verified": true,
      "phone_verified": false,
      "profile_completed": true,
      "created_at": "2025-01-20T12:00:00Z",
      "updated_at": "2025-01-20T12:00:00Z"
    },
    "message": "获取成功"
  }
  ```

## 自动化功能

### 环境变量管理

集合会自动管理以下环境变量：

- `user_access_token`: 用户访问令牌（自动更新）
- `user_refresh_token`: 用户刷新令牌（自动更新）
- `user_login_code`: 登录验证码（自动保存）
- `user_id`: 用户ID（自动保存）
- `user_name`: 用户姓名（自动保存）
- `user_email`: 用户邮箱（自动保存）

### 测试脚本

每个请求都包含自动化测试脚本：

1. **状态码验证**: 验证响应状态码是否正确
2. **响应结构验证**: 验证响应数据结构是否完整
3. **环境变量自动保存**: 自动保存重要数据到环境变量
4. **响应时间检查**: 全局检查响应时间是否在合理范围内

## 使用技巧

### 1. 批量测试

使用 Postman Runner 可以批量执行所有测试：

1. 点击集合名称旁的 "Run" 按钮
2. 选择要执行的请求
3. 设置迭代次数和延迟时间
4. 点击 "Run" 开始执行

### 2. 环境切换

可以创建多个环境（开发、测试、生产）：

1. 复制现有环境
2. 修改 `base_url` 和其他配置
3. 在不同环境间快速切换

### 3. 数据驱动测试

可以使用 CSV 文件进行数据驱动测试：

1. 创建包含测试数据的 CSV 文件
2. 在 Runner 中上传 CSV 文件
3. 使用 `{{column_name}}` 引用 CSV 数据

## 故障排除

### 常见问题

1. **401 未授权错误**
   - 检查 `user_access_token` 是否正确设置
   - 确认令牌是否已过期，尝试刷新令牌

2. **404 路径不存在**
   - 检查 `base_url` 是否正确
   - 确认 API 服务器是否正在运行

3. **验证码错误**
   - 确保在登录第一步后立即执行第二步
   - 检查验证码是否已过期（通常5分钟有效期）

4. **环境变量未保存**
   - 检查测试脚本是否正确执行
   - 确认响应状态码是否为预期值

### 调试技巧

1. **查看控制台日志**
   - 打开 Postman Console（View → Show Postman Console）
   - 查看详细的请求和响应信息

2. **检查环境变量**
   - 点击环境名称查看当前变量值
   - 确认自动保存的变量是否正确

3. **手动设置变量**
   - 如果自动保存失败，可以手动复制响应中的值
   - 在环境变量中手动设置所需的值

## 注意事项

1. **安全性**: 不要在生产环境中使用默认的测试账户信息
2. **令牌有效期**: 访问令牌有一定的有效期，过期后需要使用刷新令牌获取新的访问令牌
3. **验证码有效期**: 登录验证码通常有5分钟的有效期
4. **并发限制**: 注意 API 的速率限制，避免过于频繁的请求
5. **数据清理**: 测试完成后，建议清理测试数据

## 支持

如果在使用过程中遇到问题，请：

1. 检查 API 服务器日志
2. 查看 Postman Console 输出
3. 确认环境变量配置
4. 联系开发团队获取支持