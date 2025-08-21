# Trusioo Verification API Postman 集合

这是 Trusioo 验证码模块的 Postman API 测试集合，包含发送验证码和验证验证码的完整功能。

## 文件结构

```
Postman/auth/
├── Trusioo_Verification.postman_collection.json    # 验证码模块 API 集合
├── Trusioo_Verification.postman_environment.json   # 环境变量配置
└── Verification_README.md                           # 使用说明文档
```

## 快速开始

### 1. 导入集合和环境

1. 打开 Postman
2. 点击 "Import" 按钮
3. 导入 `Trusioo_Verification.postman_collection.json`
4. 导入 `Trusioo_Verification.postman_environment.json`
5. 在右上角选择 "Trusioo Verification Environment" 环境

### 2. 配置环境变量

在环境变量中配置以下参数：

| 变量名 | 描述 | 示例值 |
|--------|------|--------|
| `base_url` | API 基础地址 | `http://localhost:8080` |
| `verification_email` | 测试邮箱地址 | `test@example.com` |
| `verification_phone` | 测试手机号码 | `+1234567890` |
| `verification_code` | 验证码（手动输入） | `123456` |
| `verification_type` | 验证类型 | `register/login/reset_password/forgot_password/activate` |
| `verification_target` | 验证目标类型 | `email/phone` |

## 推荐测试执行流程

### 基本验证码流程

1. **发送验证码**
   - 选择对应场景的发送验证码接口
   - 执行请求，系统会发送验证码到指定邮箱
   - 自动保存过期时间到环境变量

2. **验证验证码**
   - 手动设置 `verification_code` 环境变量为收到的验证码
   - 执行对应场景的验证接口
   - 自动验证验证码有效性

### 完整测试场景

#### 注册场景
1. 发送验证码 - 注册场景
2. 验证验证码 - 注册场景

#### 登录场景
1. 发送验证码 - 登录场景
2. 验证验证码 - 登录场景

#### 密码重置场景
1. 发送验证码 - 重置密码场景
2. 验证验证码 - 重置密码场景

#### 忘记密码场景
1. 发送验证码 - 忘记密码场景
2. 验证验证码 - 忘记密码场景

#### 账户激活场景
1. 发送验证码 - 激活账户场景
2. 验证验证码 - 激活账户场景

## API 接口详细说明

### 验证码管理

#### 1. 发送验证码 (POST /api/v1/verification/send)

**请求参数：**
```json
{
  "target": "test@example.com",     // 目标邮箱或手机号
  "type": "register"               // 验证类型
}
```

**验证类型说明：**
- `register`: 注册验证码
- `login`: 登录验证码
- `reset_password`: 重置密码验证码
- `forgot_password`: 忘记密码验证码
- `activate`: 激活账户验证码

**响应示例：**
```json
{
  "data": {
    "message": "验证码已发送",
    "expired_at": "2025-01-20T12:05:00Z"
  },
  "message": "验证码发送成功"
}
```

**自动化功能：**
- 自动保存过期时间到 `verification_expired_at` 环境变量
- 验证响应状态码和数据结构
- 检查响应时间（< 5秒）

#### 2. 验证验证码 (POST /api/v1/verification/verify)

**请求参数：**
```json
{
  "target": "test@example.com",     // 目标邮箱或手机号
  "type": "register",              // 验证类型
  "code": "123456"                 // 验证码
}
```

**响应示例：**
```json
{
  "data": {
    "message": "验证码验证成功",
    "valid": true
  },
  "message": "验证成功"
}
```

**自动化功能：**
- 自动保存验证结果到 `verification_valid` 环境变量
- 验证响应状态码和数据结构
- 检查验证码有效性
- 检查响应时间（< 5秒）

## 自动化功能

### 环境变量管理

集合会自动管理以下环境变量：

- `verification_expired_at`: 验证码过期时间
- `verification_valid`: 验证码验证结果

### 测试脚本功能

每个请求都包含自动化测试脚本：

1. **状态码验证**: 确保返回 200 状态码
2. **响应结构验证**: 检查必需的响应字段
3. **数据有效性验证**: 验证关键数据的正确性
4. **性能测试**: 检查响应时间
5. **环境变量自动更新**: 保存重要数据供后续请求使用

### 全局脚本

- **预请求脚本**: 记录请求信息
- **测试脚本**: 通用性能检查

## 使用技巧

### 1. 批量测试

使用 Postman Runner 可以批量执行所有验证码相关的测试：

1. 点击集合右侧的 "Run" 按钮
2. 选择要执行的请求
3. 设置迭代次数和延迟
4. 点击 "Run" 开始执行

### 2. 环境切换

可以创建多个环境用于不同的测试场景：

- **开发环境**: `http://localhost:8080`
- **测试环境**: `https://test-api.trusioo.com`
- **生产环境**: `https://api.trusioo.com`

### 3. 数据驱动测试

可以使用 CSV 文件进行数据驱动测试：

```csv
verification_email,verification_type
test1@example.com,register
test2@example.com,login
test3@example.com,reset_password
```

### 4. 监控和报告

使用 Postman Monitor 可以定期执行测试并生成报告：

1. 在集合页面点击 "Monitor"
2. 配置监控频率和通知
3. 查看历史执行结果和趋势

## 故障排除

### 常见问题

1. **验证码发送失败**
   - 检查邮箱地址格式是否正确
   - 确认服务器邮件服务配置正常
   - 检查网络连接

2. **验证码验证失败**
   - 确认验证码输入正确
   - 检查验证码是否已过期
   - 确认验证类型匹配

3. **环境变量未更新**
   - 检查测试脚本是否正确执行
   - 确认响应数据格式正确
   - 查看 Postman Console 中的错误信息

4. **请求超时**
   - 检查服务器状态
   - 确认网络连接稳定
   - 适当增加超时时间

### 调试技巧

1. **查看 Console**: 在 Postman 底部打开 Console 查看详细日志
2. **检查环境变量**: 在环境管理器中查看变量值
3. **使用断点**: 在测试脚本中添加 `console.log()` 进行调试
4. **网络检查**: 使用 Postman 的网络面板查看请求详情

## 扩展功能

### 1. 添加新的验证类型

如需添加新的验证类型，可以：

1. 复制现有的请求
2. 修改请求体中的 `type` 字段
3. 更新请求名称和描述
4. 调整测试脚本（如需要）

### 2. 集成 CI/CD

可以使用 Newman（Postman 的命令行工具）在 CI/CD 流水线中执行测试：

```bash
newman run Trusioo_Verification.postman_collection.json \
  -e Trusioo_Verification.postman_environment.json \
  --reporters cli,json
```

### 3. 性能测试

可以配置更严格的性能测试：

```javascript
pm.test('Response time is acceptable', function () {
    pm.expect(pm.response.responseTime).to.be.below(1000); // 1秒内
});
```

## 版本历史

- **v1.0.0**: 初始版本，包含基本的验证码发送和验证功能
- 支持多种验证场景（注册、登录、重置密码、忘记密码、激活账户）
- 完整的自动化测试脚本
- 环境变量自动管理

## 技术支持

如有问题或建议，请联系开发团队或在项目仓库中提交 Issue。