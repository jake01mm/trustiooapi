# Trusioo Auth API Postman Collection

这是 Trusioo 认证模块的完整 Postman 测试集合，包含用户认证、管理员认证和验证码功能的所有 API 端点。

## 📁 文件说明

- `Trusioo_Auth_Complete_Collection.json` - 完整的 Postman 集合文件
- `Trusioo_Auth_Environment.json` - 环境变量配置文件
- `README.md` - 使用说明文档

## 🚀 快速开始

### 1. 导入集合和环境

1. 打开 Postman
2. 点击 "Import" 按钮
3. 选择 `Trusioo_Auth_Complete_Collection.json` 文件导入集合
4. 再次点击 "Import" 导入 `Trusioo_Auth_Environment.json` 环境文件
5. 在右上角选择 "Trusioo Auth Environment" 环境

### 2. 配置环境变量

在环境变量中配置以下参数：

- `baseUrl`: API 基础 URL（默认：http://localhost:8080/api/v1）
- `testUserEmail`: 测试用户邮箱
- `testUserPassword`: 测试用户密码
- `testAdminEmail`: 测试管理员邮箱
- `testAdminPassword`: 测试管理员密码

### 3. 运行测试

建议按以下顺序执行测试：

1. **用户认证流程**：
   - 用户注册
   - 用户登录（发送验证码）
   - 用户登录（验证码验证）
   - 获取用户资料
   - 刷新访问令牌

2. **管理员认证流程**：
   - 管理员登录（发送验证码）
   - 管理员登录（验证码验证）
   - 获取管理员资料
   - 刷新管理员令牌

3. **密码重置功能**：
   - 忘记密码（发送重置码）
   - 重置密码

4. **错误场景测试**：
   - 无效邮箱格式
   - 弱密码测试
   - 无效令牌测试
   - 过期验证码测试

## 📋 API 端点列表

### 🔐 用户认证 (User Authentication)

| 端点 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 用户注册 | POST | `/auth/register` | 注册新用户 |
| 用户登录 - 发送验证码 | POST | `/auth/login` | 发送登录验证码 |
| 用户登录 - 验证码验证 | POST | `/auth/verify-login` | 验证登录验证码 |
| 忘记密码 - 发送重置码 | POST | `/auth/forgot-password` | 发送密码重置验证码 |
| 重置密码 | POST | `/auth/reset-password` | 重置用户密码 |
| 刷新访问令牌 | POST | `/auth/refresh` | 刷新用户访问令牌 |
| 获取用户资料 | GET | `/auth/profile` | 获取当前用户信息 |

### 👨‍💼 管理员认证 (Admin Authentication)

| 端点 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 管理员登录 - 发送验证码 | POST | `/admin/login` | 发送管理员登录验证码 |
| 管理员登录 - 验证码验证 | POST | `/admin/verify-login` | 验证管理员登录验证码 |
| 管理员刷新访问令牌 | POST | `/admin/refresh` | 刷新管理员访问令牌 |
| 获取管理员资料 | GET | `/admin/profile` | 获取当前管理员信息 |

### 🔄 密码重置功能 (Password Reset)

| 端点 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 忘记密码 | POST | `/user/forgot-password` | 发送密码重置验证码 |
| 重置密码 | POST | `/user/reset-password` | 使用验证码重置密码 |

> **注意**: 验证码功能是内部服务模块，通过其他接口间接调用，不提供独立的HTTP接口。

## 🔧 环境变量说明

### 基础配置
- `baseUrl`: API 基础 URL
- `prodBaseUrl`: 生产环境 API URL（备用）

### 测试数据
- `testUserEmail`: 测试用户邮箱
- `testUserPassword`: 测试用户密码
- `testAdminEmail`: 测试管理员邮箱
- `testAdminPassword`: 测试管理员密码
- `verificationCode`: 测试验证码（实际使用时需要从邮件获取）

### 动态令牌（自动设置）
- `userAccessToken`: 用户访问令牌
- `userRefreshToken`: 用户刷新令牌
- `adminAccessToken`: 管理员访问令牌
- `adminRefreshToken`: 管理员刷新令牌

## 🧪 测试脚本功能

### 预请求脚本
- 自动添加请求 ID 头
- 设置时间戳变量

### 测试脚本
- 响应时间验证（< 5秒）
- 响应头验证
- JSON 格式验证
- 状态码验证
- 自动保存令牌到环境变量

## 📝 使用注意事项

1. **验证码获取**：实际测试时，需要从邮件中获取真实的验证码，不能使用默认的 `123456`

2. **令牌管理**：登录成功后，访问令牌和刷新令牌会自动保存到环境变量中

3. **环境切换**：可以通过修改 `baseUrl` 在开发环境和生产环境之间切换

4. **错误测试**：错误场景测试用于验证 API 的错误处理机制

5. **顺序执行**：某些测试依赖于前面的测试结果（如令牌），建议按顺序执行

## 🔍 故障排除

### 常见问题

1. **401 Unauthorized**
   - 检查访问令牌是否有效
   - 确认是否已正确登录
   - 验证令牌是否已过期

2. **422 Validation Error**
   - 检查请求参数格式
   - 确认必填字段是否完整
   - 验证邮箱格式是否正确

3. **429 Too Many Requests**
   - 等待一段时间后重试
   - 检查验证码发送频率限制

4. **500 Internal Server Error**
   - 检查服务器是否正常运行
   - 查看服务器日志获取详细错误信息

### 调试技巧

1. 使用 Postman Console 查看详细的请求和响应信息
2. 检查环境变量是否正确设置
3. 确认 API 服务器地址和端口
4. 验证请求体 JSON 格式是否正确

## 📞 支持

如有问题或建议，请联系开发团队或在项目仓库中提交 Issue。