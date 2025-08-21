# Admin Auth API 文档

## 概述

这是 Trusioo API 的管理员认证和用户管理模块的前端开发文档。

## 基础信息

- **API Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token
- **响应格式**: 统一 JSON 格式

## 通用响应格式

所有 API 响应都遵循以下格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {} // 具体数据，可选
}
```

### 状态码说明

- `200`: 成功
- `400`: 参数错误/验证失败
- `401`: 未授权/认证失败
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

## 认证流程

### 1. 管理员登录（两步验证）

#### 第一步：发送登录验证码

**POST** `/admin/auth/login`

**请求参数:**
```json
{
  "email": "admin@trusioo.com",
  "password": "admin123"
}
```

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Login verification code sent successfully",
    "login_code": "123456",
    "expires_in": 300
  }
}
```

#### 第二步：验证登录验证码

**POST** `/admin/auth/login/verify`

**请求参数:**
```json
{
  "email": "admin@trusioo.com",
  "code": "123456"
}
```

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200,
    "token_type": "Bearer",
    "admin": {
      "id": 1,
      "name": "管理员",
      "email": "admin@trusioo.com",
      "phone": null,
      "image_key": "",
      "role": "admin",
      "is_super": true,
      "status": "active",
      "last_login_at": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    "login_session": {
      "ip": "127.0.0.1",
      "country": "China",
      "city": "Beijing",
      "region": "Beijing",
      "timezone": "Asia/Shanghai",
      "organization": "Local Network",
      "location": "39.9042,116.4074",
      "is_trusted": true
    }
  }
}
```

### 2. 刷新访问令牌

**POST** `/admin/auth/refresh`

**请求参数:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200,
    "token_type": "Bearer",
    "admin": {
      // 管理员信息
    }
  }
}
```

### 3. 忘记密码流程

#### 第一步：发送重置密码验证码

**POST** `/admin/auth/forgot-password`

**请求参数:**
```json
{
  "email": "admin@trusioo.com"
}
```

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Password reset code sent successfully",
    "expires_in": 300
  }
}
```

#### 第二步：重置密码

**POST** `/admin/auth/reset-password`

**请求参数:**
```json
{
  "email": "admin@trusioo.com",
  "code": "123456",
  "password": "newpassword123"
}
```

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Password reset successfully"
  }
}
```

## 管理员功能 API

> 以下 API 需要在请求头中携带访问令牌：
> `Authorization: Bearer {access_token}`

### 1. 获取管理员资料

**GET** `/admin/profile`

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "管理员",
    "email": "admin@trusioo.com",
    "phone": null,
    "image_key": "",
    "role": "admin",
    "is_super": true,
    "status": "active",
    "last_login_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### 2. 获取用户统计

**GET** `/admin/users/stats`

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_users": 1250,
    "active_users": 1180,
    "inactive_users": 70,
    "registered_today": 15,
    "registered_this_week": 89,
    "registered_this_month": 342
  }
}
```

### 3. 获取用户列表

**GET** `/admin/users`

**查询参数:**
- `page`: 页码 (默认: 1)
- `page_size`: 每页条数 (默认: 20, 最大: 100)
- `status`: 用户状态筛选 (`active`, `inactive`, `all`)
- `email`: 邮箱筛选
- `phone`: 手机号筛选

**示例请求:**
```
GET /admin/users?page=1&page_size=20&status=active
```

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 1180,
    "page": 1,
    "size": 20,
    "users": [
      {
        "id": 1,
        "name": "张三",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "image_key": "",
        "status": "active",
        "email_verified": true,
        "phone_verified": true,
        "auto_registered": false,
        "profile_completed": true,
        "last_login_at": "2024-01-01T12:00:00Z",
        "created_at": "2024-01-01T00:00:00Z"
      }
      // ... 更多用户
    ]
  }
}
```

### 4. 获取用户详情

**GET** `/admin/users/{id}`

**路径参数:**
- `id`: 用户ID

**成功响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "image_key": "",
    "status": "active",
    "email_verified": true,
    "phone_verified": true,
    "auto_registered": false,
    "profile_completed": true,
    "last_login_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## 错误处理

### 常见错误响应

**参数验证错误 (400):**
```json
{
  "code": 400,
  "message": "Email or password is incorrect"
}
```

**未授权 (401):**
```json
{
  "code": 401,
  "message": "Invalid refresh token"
}
```

**资源不存在 (404):**
```json
{
  "code": 404,
  "message": "User not found"
}
```

**服务器错误 (500):**
```json
{
  "code": 500,
  "message": "Internal server error"
}
```

## 前端实现建议

### 1. 认证状态管理

```javascript
// 推荐使用 Zustand 或 Redux Toolkit
const useAuthStore = create((set, get) => ({
  admin: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
  
  login: async (email, password) => {
    // 第一步：发送验证码
    const codeResponse = await api.post('/admin/auth/login', { email, password });
    return codeResponse.data;
  },
  
  verifyLogin: async (email, code) => {
    // 第二步：验证登录
    const response = await api.post('/admin/auth/login/verify', { email, code });
    const { access_token, refresh_token, admin } = response.data.data;
    
    set({
      admin,
      accessToken: access_token,
      refreshToken: refresh_token,
      isAuthenticated: true
    });
    
    // 存储到 localStorage
    localStorage.setItem('admin_access_token', access_token);
    localStorage.setItem('admin_refresh_token', refresh_token);
  },
  
  logout: () => {
    set({
      admin: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false
    });
    localStorage.removeItem('admin_access_token');
    localStorage.removeItem('admin_refresh_token');
  }
}));
```

### 2. HTTP 客户端配置

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器 - 自动添加认证头
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器 - 处理 token 过期
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Token 过期，尝试刷新
      const refreshToken = localStorage.getItem('admin_refresh_token');
      if (refreshToken) {
        try {
          const response = await axios.post('/admin/auth/refresh', {
            refresh_token: refreshToken
          });
          const { access_token } = response.data.data;
          localStorage.setItem('admin_access_token', access_token);
          
          // 重试原请求
          error.config.headers.Authorization = `Bearer ${access_token}`;
          return axios.request(error.config);
        } catch (refreshError) {
          // 刷新失败，跳转到登录页
          window.location.href = '/admin/login';
        }
      }
    }
    return Promise.reject(error);
  }
);
```

### 3. 路由保护

```javascript
// React Router 示例
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  
  if (!isAuthenticated) {
    return <Navigate to="/admin/login" replace />;
  }
  
  return children;
};
```

## 开发环境配置

### 环境变量

```env
# .env.local
REACT_APP_API_BASE_URL=http://localhost:8080/api/v1
REACT_APP_ADMIN_URL=http://localhost:3001
```

### CORS 配置

后端已配置 CORS，允许以下域名：
- `http://localhost:3000` (用户端)
- `http://localhost:3001` (管理端)

## 默认管理员账户

**测试账户:**
- 邮箱: `admin@trusioo.com`
- 密码: `admin123`

## 注意事项

1. **Token 管理**: 访问令牌有效期 2 小时，刷新令牌有效期 7 天
2. **安全性**: 生产环境请更换默认管理员密码
3. **验证码**: 开发环境验证码会在响应中返回，生产环境需要邮件接收
4. **分页**: 用户列表支持分页，建议每页不超过 100 条
5. **状态筛选**: 用户状态包括 `active`、`inactive`
6. **错误处理**: 请根据 HTTP 状态码和响应消息进行适当的错误提示

## 联系方式

如有 API 相关问题，请联系后端开发团队。