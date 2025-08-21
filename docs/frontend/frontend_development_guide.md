# 前端开发快速启动指南

## 项目概述

Trusioo API 是一个前后端分离的项目，后端提供 RESTful API，前端需要开发管理后台和用户端应用。

## 后端服务信息

### 服务地址
- **开发环境**: `http://localhost:8080`
- **API Base URL**: `http://localhost:8080/api/v1`

### 前端应用配置
根据后端配置，前端应用应该运行在：
- **用户端**: `http://localhost:3000`
- **管理端**: `http://localhost:3001`

## 技术栈建议

### 管理后台 (推荐)
```json
{
  "框架": "React 18 + TypeScript",
  "UI库": "Ant Design 5.x",
  "状态管理": "Zustand / Redux Toolkit",
  "路由": "React Router v6",
  "HTTP客户端": "Axios",
  "构建工具": "Vite",
  "样式": "Tailwind CSS + Ant Design"
}
```

### 用户端 (推荐)
```json
{
  "框架": "React 18 + TypeScript / Vue 3 + TypeScript",
  "UI库": "Tailwind CSS + Headless UI",
  "状态管理": "Zustand / Pinia",
  "路由": "React Router v6 / Vue Router 4",
  "HTTP客户端": "Axios",
  "构建工具": "Vite"
}
```

## 快速开始

### 1. 创建管理后台项目

```bash
# 使用 Vite 创建 React + TypeScript 项目
npm create vite@latest trusioo-admin -- --template react-ts
cd trusioo-admin

# 安装依赖
npm install

# 安装 UI 库和工具
npm install antd @ant-design/icons
npm install axios zustand
npm install react-router-dom
npm install @types/node

# 安装开发依赖
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

### 2. 项目结构建议

```
trusioo-admin/
├── src/
│   ├── components/          # 通用组件
│   │   ├── Layout/         # 布局组件
│   │   ├── Auth/           # 认证相关组件
│   │   └── Common/         # 通用组件
│   ├── pages/              # 页面组件
│   │   ├── Login/          # 登录页面
│   │   ├── Dashboard/      # 仪表板
│   │   ├── Users/          # 用户管理
│   │   └── Profile/        # 个人资料
│   ├── stores/             # 状态管理
│   │   ├── authStore.ts    # 认证状态
│   │   └── userStore.ts    # 用户数据
│   ├── services/           # API 服务
│   │   ├── api.ts          # HTTP 客户端配置
│   │   ├── authService.ts  # 认证 API
│   │   └── userService.ts  # 用户管理 API
│   ├── types/              # TypeScript 类型定义
│   │   ├── auth.ts         # 认证相关类型
│   │   └── user.ts         # 用户相关类型
│   ├── utils/              # 工具函数
│   │   ├── constants.ts    # 常量定义
│   │   └── helpers.ts      # 辅助函数
│   └── App.tsx
├── public/
└── package.json
```

### 3. 环境配置

创建 `.env.local` 文件：

```env
# API 配置
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=Trusioo 管理后台

# 开发配置
VITE_DEV_MODE=true
```

### 4. 核心代码示例

#### API 客户端配置 (`src/services/api.ts`)

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const refreshToken = localStorage.getItem('admin_refresh_token');
      if (refreshToken) {
        try {
          const response = await axios.post(`${import.meta.env.VITE_API_BASE_URL}/admin/auth/refresh`, {
            refresh_token: refreshToken
          });
          
          const { access_token } = response.data.data;
          localStorage.setItem('admin_access_token', access_token);
          
          // 重试原请求
          error.config.headers.Authorization = `Bearer ${access_token}`;
          return axios.request(error.config);
        } catch (refreshError) {
          localStorage.removeItem('admin_access_token');
          localStorage.removeItem('admin_refresh_token');
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

export default api;
```

#### 认证状态管理 (`src/stores/authStore.ts`)

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../services/api';

interface Admin {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  is_super: boolean;
  status: string;
}

interface AuthState {
  admin: Admin | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  
  // Actions
  login: (email: string, password: string) => Promise<{ login_code: string; expires_in: number }>;
  verifyLogin: (email: string, code: string) => Promise<void>;
  logout: () => void;
  refreshAccessToken: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      admin: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,

      login: async (email: string, password: string) => {
        const response = await api.post('/admin/auth/login', { email, password });
        return response.data.data;
      },

      verifyLogin: async (email: string, code: string) => {
        const response = await api.post('/admin/auth/login/verify', { email, code });
        const { access_token, refresh_token, admin } = response.data.data;
        
        set({
          admin,
          accessToken: access_token,
          refreshToken: refresh_token,
          isAuthenticated: true,
        });
        
        localStorage.setItem('admin_access_token', access_token);
        localStorage.setItem('admin_refresh_token', refresh_token);
      },

      logout: () => {
        set({
          admin: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
        });
        localStorage.removeItem('admin_access_token');
        localStorage.removeItem('admin_refresh_token');
      },

      refreshAccessToken: async () => {
        const { refreshToken } = get();
        if (!refreshToken) throw new Error('No refresh token');
        
        const response = await api.post('/admin/auth/refresh', {
          refresh_token: refreshToken
        });
        
        const { access_token, refresh_token: new_refresh_token } = response.data.data;
        
        set({
          accessToken: access_token,
          refreshToken: new_refresh_token,
        });
        
        localStorage.setItem('admin_access_token', access_token);
        localStorage.setItem('admin_refresh_token', new_refresh_token);
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        admin: state.admin,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
```

#### 登录页面组件 (`src/pages/Login/LoginPage.tsx`)

```typescript
import React, { useState } from 'react';
import { Form, Input, Button, Card, message, Steps } from 'antd';
import { UserOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../stores/authStore';
import { useNavigate } from 'react-router-dom';

const LoginPage: React.FC = () => {
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [email, setEmail] = useState('');
  const { login, verifyLogin } = useAuthStore();
  const navigate = useNavigate();

  const handleLogin = async (values: { email: string; password: string }) => {
    setLoading(true);
    try {
      const result = await login(values.email, values.password);
      setEmail(values.email);
      setCurrentStep(1);
      message.success(`验证码已发送，有效期 ${result.expires_in / 60} 分钟`);
    } catch (error: any) {
      message.error(error.response?.data?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyCode = async (values: { code: string }) => {
    setLoading(true);
    try {
      await verifyLogin(email, values.code);
      message.success('登录成功');
      navigate('/dashboard');
    } catch (error: any) {
      message.error(error.response?.data?.message || '验证失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold">Trusioo 管理后台</h1>
        </div>
        
        <Steps current={currentStep} className="mb-6">
          <Steps.Step title="账号验证" icon={<UserOutlined />} />
          <Steps.Step title="安全验证" icon={<SafetyOutlined />} />
        </Steps>

        {currentStep === 0 && (
          <Form onFinish={handleLogin} layout="vertical">
            <Form.Item
              name="email"
              label="邮箱"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input 
                prefix={<UserOutlined />} 
                placeholder="admin@trusioo.com"
                size="large"
              />
            </Form.Item>
            
            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password 
                prefix={<LockOutlined />} 
                placeholder="请输入密码"
                size="large"
              />
            </Form.Item>
            
            <Form.Item>
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={loading}
                size="large"
                className="w-full"
              >
                发送验证码
              </Button>
            </Form.Item>
          </Form>
        )}

        {currentStep === 1 && (
          <Form onFinish={handleVerifyCode} layout="vertical">
            <Form.Item
              name="code"
              label="验证码"
              rules={[
                { required: true, message: '请输入验证码' },
                { len: 6, message: '验证码为6位数字' }
              ]}
            >
              <Input 
                prefix={<SafetyOutlined />} 
                placeholder="请输入6位验证码"
                size="large"
                maxLength={6}
              />
            </Form.Item>
            
            <Form.Item>
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={loading}
                size="large"
                className="w-full"
              >
                登录
              </Button>
            </Form.Item>
            
            <Form.Item>
              <Button 
                type="link" 
                onClick={() => setCurrentStep(0)}
                className="w-full"
              >
                返回重新登录
              </Button>
            </Form.Item>
          </Form>
        )}
      </Card>
    </div>
  );
};

export default LoginPage;
```

## 开发流程

### 1. 启动后端服务

```bash
cd /Users/laitsim/trusioo_api
make dev  # 或者 go run cmd/main.go
```

### 2. 启动前端开发服务器

```bash
cd trusioo-admin
npm run dev
```

### 3. 测试 API 连接

使用提供的 Postman 集合测试 API：
1. 导入 `docs/admin_auth_postman.json`
2. 设置环境变量 `baseUrl` 为 `http://localhost:8080/api/v1`
3. 测试登录流程

## 重要文件

1. **API 文档**: `docs/admin_auth_api.md` - 完整的 API 接口文档
2. **Postman 集合**: `docs/admin_auth_postman.json` - API 测试集合
3. **环境配置**: `.env.example` - 后端环境变量示例

## 默认测试账户

- **邮箱**: `admin@trusioo.com`
- **密码**: `admin123`

## 开发注意事项

1. **CORS**: 后端已配置 CORS，支持 `localhost:3000` 和 `localhost:3001`
2. **Token 管理**: 访问令牌 2 小时过期，刷新令牌 7 天过期
3. **错误处理**: 统一的错误响应格式，需要根据 code 和 message 处理
4. **分页**: 用户列表等接口支持分页，注意处理分页参数
5. **验证码**: 开发环境验证码会在 API 响应中返回，生产环境需要邮件

## 下一步

1. 完成管理后台的基础功能开发
2. 实现用户端应用
3. 集成卡片检测功能（开发中）
4. 添加更多管理功能

如有问题，请参考 API 文档或联系后端开发团队。