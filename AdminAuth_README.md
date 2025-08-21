# Trusioo Admin Auth API Postman Collection

## 概述

本 Postman 集合包含了 Trusioo 管理员认证和用户管理模块的完整 API 端点测试。集合涵盖了管理员登录、认证、用户管理等核心功能，提供了完整的测试覆盖和自动化验证。

## 文件说明

- **Trusioo_AdminAuth_Collection.postman_collection.json**: 主要的 Postman 集合文件
- **Trusioo_AdminAuth_Environment.postman_environment.json**: 环境变量配置文件
- **AdminAuth_README.md**: 本使用说明文档

## 功能模块

### 1. 管理员认证 (Admin Authentication)

#### 1.1 管理员登录 (第一步)
- **端点**: `POST /api/v1/admin/auth/login`
- **描述**: 发送管理员邮箱和密码，获取登录验证码
- **请求体**:
  ```json
  {
    "email": "{{adminEmail}}",
    "password": "{{adminPassword}}"
  }
  ```
- **成功响应**:
  ```json
  {
    "success": true,
    "data": {
      "login_code": "123456",
      "expires_at": "2024-01-20T10:05:00Z"
    }
  }
  ```

#### 1.2 管理员登录验证 (第二步)
- **端点**: `POST /api/v1/admin/auth/login/verify`
- **描述**: 使用邮箱和验证码完成登录，获取访问令牌
- **请求体**:
  ```json
  {
    "email": "{{adminEmail}}",
    "login_code": "{{adminLoginCode}}"
  }
  ```
- **成功响应**:
  ```json
  {
    "success": true,
    "data": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
      "admin": {
        "id": "admin-uuid",
        "name": "Admin User",
        "email": "admin@trusioo.com",
        "role": "admin"
      }
    }
  }
  ```

#### 1.3 刷新令牌
- **端点**: `POST /api/v1/admin/auth/refresh`
- **描述**: 使用刷新令牌获取新的访问令牌
- **请求体**:
  ```json
  {
    "refresh_token": "{{adminRefreshToken}}"
  }
  ```

#### 1.4 忘记密码 (第一步)
- **端点**: `POST /api/v1/admin/auth/forgot-password`
- **描述**: 发送重置密码请求，获取重置验证码
- **请求体**:
  ```json
  {
    "email": "{{adminEmail}}"
  }
  ```

#### 1.5 重置密码 (第二步)
- **端点**: `POST /api/v1/admin/auth/reset-password`
- **描述**: 使用验证码重置密码
- **请求体**:
  ```json
  {
    "email": "{{adminEmail}}",
    "reset_code": "{{resetCode}}",
    "new_password": "{{newPassword}}"
  }
  ```

#### 1.6 获取管理员资料
- **端点**: `GET /api/v1/admin/profile`
- **描述**: 获取当前登录管理员的详细信息
- **认证**: 需要 Bearer Token
- **成功响应**:
  ```json
  {
    "success": true,
    "data": {
      "id": "admin-uuid",
      "name": "Admin User",
      "email": "admin@trusioo.com",
      "phone": "+1234567890",
      "role": "admin",
      "is_super": false,
      "status": "active",
      "last_login_at": "2024-01-20T09:00:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 2. 用户管理 (User Management)

#### 2.1 获取用户统计
- **端点**: `GET /api/v1/admin/users/stats`
- **描述**: 获取用户统计信息
- **认证**: 需要 Bearer Token
- **成功响应**:
  ```json
  {
    "success": true,
    "data": {
      "total_users": 1500,
      "active_users": 1200,
      "inactive_users": 300,
      "registered_today": 25,
      "registered_this_week": 150,
      "registered_this_month": 600
    }
  }
  ```

#### 2.2 获取用户列表
- **端点**: `GET /api/v1/admin/users`
- **描述**: 获取分页用户列表，支持多种过滤条件
- **认证**: 需要 Bearer Token
- **查询参数**:
  - `page`: 页码 (默认: 1)
  - `page_size`: 每页数量 (默认: 20, 最大: 100)
  - `status`: 状态过滤 (active, inactive, all)
  - `email`: 邮箱过滤
  - `phone`: 电话过滤
- **成功响应**:
  ```json
  {
    "success": true,
    "data": {
      "total": 1500,
      "page": 1,
      "size": 20,
      "users": [
        {
          "id": "user-uuid",
          "name": "John Doe",
          "email": "john@example.com",
          "phone": "+1234567890",
          "status": "active",
          "email_verified": true,
          "phone_verified": false,
          "created_at": "2024-01-15T10:00:00Z"
        }
      ]
    }
  }
  ```

#### 2.3 获取用户详情
- **端点**: `GET /api/v1/admin/users/{userId}`
- **描述**: 获取指定用户的详细信息
- **认证**: 需要 Bearer Token
- **路径参数**: `userId` - 用户ID
- **成功响应**:
  ```json
  {
    "success": true,
    "data": {
      "id": "user-uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "image_key": "profile-image-key",
      "status": "active",
      "email_verified": true,
      "phone_verified": false,
      "auto_registered": false,
      "profile_completed": true,
      "last_login_at": "2024-01-20T08:30:00Z",
      "created_at": "2024-01-15T10:00:00Z"
    }
  }
  ```

## 环境变量说明

### 基础配置
- **baseUrl**: API 服务器基础 URL (默认: `http://localhost:8080`)
- **adminEmail**: 管理员邮箱 (默认: `admin@trusioo.com`)
- **adminPassword**: 管理员密码 (默认: `admin123456`)

### 自动填充变量
以下变量会在测试过程中自动填充，无需手动设置：
- **adminLoginCode**: 登录验证码
- **adminAccessToken**: 访问令牌
- **adminRefreshToken**: 刷新令牌
- **adminId**: 管理员用户ID
- **adminName**: 管理员用户名
- **resetCode**: 密码重置验证码
- **testUserId**: 测试用户ID

### 用户管理过滤参数
- **userListPage**: 用户列表页码 (默认: 1)
- **userListPageSize**: 每页数量 (默认: 20)
- **userStatusFilter**: 状态过滤 (默认: all)
- **userEmailFilter**: 邮箱过滤
- **userPhoneFilter**: 电话过滤
- **newPassword**: 重置密码用的新密码

## 使用方法

### 1. 导入集合和环境
1. 在 Postman 中导入 `Trusioo_AdminAuth_Collection.postman_collection.json`
2. 导入 `Trusioo_AdminAuth_Environment.postman_environment.json`
3. 选择 "Trusioo Admin Auth Environment" 作为活动环境

### 2. 配置环境变量
1. 根据实际情况修改 `baseUrl`
2. 设置正确的 `adminEmail` 和 `adminPassword`
3. 其他变量保持默认值或根据需要调整

### 3. 执行测试流程

#### 基本认证流程：
1. **Admin Login (Step 1)** - 获取登录验证码
2. **Admin Login Verify (Step 2)** - 完成登录获取令牌
3. **Get Admin Profile** - 验证登录状态

#### 用户管理流程：
1. 确保已完成管理员登录
2. **Get User Statistics** - 查看用户统计
3. **Get User List** - 获取用户列表
4. **Get User Detail** - 查看用户详情

#### 密码重置流程：
1. **Forgot Password (Step 1)** - 发送重置请求
2. **Reset Password (Step 2)** - 完成密码重置

### 4. 自动化测试

集合包含完整的测试脚本，会自动：
- 验证响应状态码
- 检查响应时间
- 验证响应数据结构
- 保存必要的环境变量
- 提供详细的控制台日志

## 测试覆盖

### 功能测试
- ✅ 管理员登录流程
- ✅ 令牌刷新机制
- ✅ 密码重置流程
- ✅ 管理员资料获取
- ✅ 用户统计查询
- ✅ 用户列表分页和过滤
- ✅ 用户详情查询

### 技术测试
- ✅ 响应时间验证
- ✅ 状态码检查
- ✅ JSON 格式验证
- ✅ 数据结构验证
- ✅ 认证令牌管理
- ✅ 环境变量自动化

## 错误处理

集合包含完整的错误处理机制：

### 常见错误响应
```json
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password"
  }
}
```

### 认证错误
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Access token is required"
  }
}
```

### 验证错误
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": {
      "email": "Email is required"
    }
  }
}
```

## 注意事项

1. **认证顺序**: 必须先完成管理员登录流程才能访问受保护的端点
2. **令牌管理**: 访问令牌会自动保存到环境变量中，无需手动复制
3. **测试数据**: 使用测试环境时请确保不会影响生产数据
4. **并发限制**: 注意 API 的并发请求限制
5. **令牌过期**: 访问令牌过期时使用刷新令牌获取新的访问令牌

## 技术支持

如有问题或建议，请联系开发团队或查看项目文档。

---

**版本**: 1.0.0  
**更新日期**: 2024-01-20  
**维护者**: Trusioo 开发团队