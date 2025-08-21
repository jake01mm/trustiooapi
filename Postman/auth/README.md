# Trusioo Admin Auth API Postman 集合

这个 Postman 集合包含了 Trusioo 项目中管理员认证和用户管理模块的所有 API 接口测试。

## 📁 文件结构

```
/Users/laitsim/trusioo_api/Postman/auth/
├── Trusioo_Admin_Auth.postman_collection.json  # 主要的 API 集合文件
├── Trusioo_Admin_Auth.postman_environment.json # 环境变量配置文件
└── README.md                                    # 使用说明文档
```

## 🚀 快速开始

### 1. 导入 Postman 集合

1. 打开 Postman 应用
2. 点击 "Import" 按钮
3. 选择 `Trusioo_Admin_Auth.postman_collection.json` 文件
4. 导入 `Trusioo_Admin_Auth.postman_environment.json` 环境文件
5. 在右上角选择 "Trusioo Admin Auth Environment" 环境

### 2. 配置环境变量

在环境变量中配置以下参数：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `base_url` | `http://localhost:8080` | API 服务器地址 |
| `admin_email` | `admin@example.com` | 管理员邮箱 |
| `admin_password` | `admin123` | 管理员密码 |
| `user_id` | `1` | 用户ID（用于用户详情查询） |

### 3. 执行测试流程

**推荐执行顺序：**

1. **管理员登录 - 第一步**：验证邮箱密码，获取验证码
2. **管理员登录 - 第二步**：使用验证码完成登录，获取 token
3. **获取管理员个人信息**：验证登录状态
4. **获取用户统计信息**：测试用户管理功能
5. **获取用户列表**：查看用户列表
6. **获取用户详情**：查看特定用户信息
7. **刷新访问令牌**：测试 token 刷新功能

## 📋 API 接口详情

### 管理员认证模块

#### 1. 管理员登录 - 第一步
- **路径**: `POST /api/v1/admin/auth/login`
- **功能**: 验证管理员邮箱和密码，发送验证码
- **请求参数**:
  ```json
  {
    "email": "admin@example.com",
    "password": "admin123"
  }
  ```
- **响应数据**:
  ```json
  {
    "data": {
      "login_code": "123456",
      "expires_in": 300
    }
  }
  ```
- **自动化功能**: 自动保存 `login_code` 到环境变量

#### 2. 管理员登录 - 第二步
- **路径**: `POST /api/v1/admin/auth/login/verify`
- **功能**: 使用验证码完成登录验证
- **请求参数**:
  ```json
  {
    "email": "admin@example.com",
    "code": "123456"
  }
  ```
- **响应数据**:
  ```json
  {
    "data": {
      "access_token": "eyJ...",
      "refresh_token": "eyJ...",
      "admin": {
        "id": 1,
        "name": "Admin",
        "email": "admin@example.com"
      }
    }
  }
  ```
- **自动化功能**: 自动保存 `access_token` 和 `refresh_token`

#### 3. 刷新访问令牌
- **路径**: `POST /api/v1/admin/auth/refresh`
- **功能**: 刷新过期的访问令牌
- **请求参数**:
  ```json
  {
    "refresh_token": "eyJ..."
  }
  ```
- **自动化功能**: 自动更新环境变量中的 token

### 管理员个人信息模块

#### 4. 获取管理员个人信息
- **路径**: `GET /api/v1/admin/profile`
- **功能**: 获取当前登录管理员的个人信息
- **请求头**: `Authorization: Bearer {access_token}`
- **响应数据**:
  ```json
  {
    "data": {
      "id": 1,
      "name": "Admin",
      "email": "admin@example.com",
      "role": "admin",
      "status": "active"
    }
  }
  ```

### 用户管理模块

#### 5. 获取用户统计信息
- **路径**: `GET /api/v1/admin/users/stats`
- **功能**: 获取用户统计数据
- **响应数据**:
  ```json
  {
    "data": {
      "total_users": 100,
      "active_users": 85,
      "inactive_users": 15
    }
  }
  ```

#### 6. 获取用户列表
- **路径**: `GET /api/v1/admin/users`
- **功能**: 分页获取用户列表
- **查询参数**:
  - `page`: 页码（默认：1）
  - `page_size`: 每页数量（默认：20）
  - `status`: 用户状态过滤（all/active/inactive）
- **响应数据**:
  ```json
  {
    "data": {
      "users": [...],
      "total": 100,
      "page": 1,
      "page_size": 20
    }
  }
  ```

#### 7. 获取用户详情
- **路径**: `GET /api/v1/admin/users/{user_id}`
- **功能**: 获取指定用户的详细信息
- **路径参数**: `user_id` - 用户ID
- **响应数据**:
  ```json
  {
    "data": {
      "id": 1,
      "email": "user@example.com",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

## 🔧 自动化功能

### 环境变量自动管理
- **Token 自动保存**: 登录成功后自动保存 `access_token` 和 `refresh_token`
- **验证码自动保存**: 第一步登录后自动保存 `login_code`
- **Token 自动刷新**: 提供刷新令牌的接口和自动更新功能

### 测试脚本
每个请求都包含自动化测试脚本：
- **状态码验证**: 验证响应状态码为 200
- **响应结构验证**: 验证响应数据包含必要字段
- **数据类型验证**: 验证关键字段的数据类型

### 全局脚本
- **预请求脚本**: 检查 token 是否存在
- **测试脚本**: 检查 401 错误并提示刷新 token

## 🛠️ 使用技巧

### 1. 批量执行测试
1. 选择整个集合或特定文件夹
2. 点击 "Run" 按钮
3. 配置运行参数
4. 查看测试报告

### 2. 环境切换
- **开发环境**: `http://localhost:8080`
- **测试环境**: `https://test-api.trusioo.com`
- **生产环境**: `https://api.trusioo.com`

### 3. 调试技巧
- 使用 Console 查看自动化脚本的日志输出
- 检查环境变量是否正确设置
- 验证请求头和请求体格式

## ⚠️ 注意事项

1. **Token 有效期**: Access token 有效期较短，请及时刷新
2. **验证码有效期**: 登录验证码有效期为 5 分钟
3. **环境变量**: 确保在正确的环境中执行测试
4. **权限验证**: 某些接口需要特定的管理员权限
5. **数据安全**: 不要在生产环境中使用测试数据

## 🔍 故障排除

### 常见问题

1. **401 Unauthorized**
   - 检查 token 是否过期
   - 尝试重新登录获取新 token
   - 验证 Authorization 头格式

2. **404 Not Found**
   - 检查 API 路径是否正确
   - 确认服务器是否正在运行
   - 验证 base_url 配置

3. **400 Bad Request**
   - 检查请求参数格式
   - 验证必填字段是否提供
   - 确认数据类型是否正确

### 联系支持
如果遇到问题，请联系开发团队或查看项目文档。

---

**最后更新**: 2024-01-20  
**版本**: v1.0.0  
**维护者**: Trusioo 开发团队