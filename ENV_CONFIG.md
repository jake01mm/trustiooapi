# 🔧 Trusioo API 环境配置说明

## 📋 配置文件概览

你的 `.env` 文件已经完善，包含了所有必要的配置项。以下是详细的配置说明：

## 🗂️ 配置分类

### 1. 数据库配置
```bash
# === 数据库配置 - Neon Database ===
DB_HOST=ep-icy-wind-a1yyruhk-pooler.ap-southeast-1.aws.neon.tech
DB_PORT=5432
DB_USER=neondb_owner
DB_PASSWORD=npg_6wOro0ezdFBa
DB_NAME=neondb

# 数据库连接池配置
DB_MAX_OPEN_CONNS=25      # 最大开放连接数
DB_MAX_IDLE_CONNS=5       # 最大空闲连接数
DB_CONN_MAX_LIFETIME=300  # 连接最大生命周期（秒）
```

### 2. JWT 安全配置
```bash
# === JWT 配置 ===
JWT_SECRET=trusioo_super_secret_jwt_key_2024_enhanced_security
JWT_REFRESH_SECRET=trusioo_super_secret_refresh_key_2024_enhanced_security
JWT_ACCESS_EXPIRE=7200    # 访问令牌过期时间（2小时）
JWT_REFRESH_EXPIRE=604800 # 刷新令牌过期时间（7天）
```
> ⚠️ **生产环境建议**: 使用更长的随机密钥（建议64位以上）

### 3. 服务器配置
```bash
# === 服务器配置 ===
PORT=8080                 # 服务器端口
ENV=development          # 运行环境：development | production
```

### 4. 跨域资源共享 (CORS)
```bash
# === CORS 配置 ===
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOW_ALL=false     # 是否允许所有来源
```

### 5. 前端 URL 配置
```bash
# === 前端 URL 配置 ===
FRONTEND_APP_URL=http://localhost:3000    # 用户端前端地址
FRONTEND_ADMIN_URL=http://localhost:3001  # 管理端前端地址
```

### 6. 日志系统配置
```bash
# === 日志配置 ===
LOG_LEVEL=info           # 日志级别：debug | info | warn | error
LOG_FORMAT=json          # 日志格式：json | text
```

### 7. 安全配置
```bash
# === 安全配置 ===
RATE_LIMIT_ENABLED=true          # 是否启用速率限制
RATE_LIMIT_REQUESTS=100          # 全局速率限制（每分钟请求数）
RATE_LIMIT_WINDOW=60             # 速率限制窗口（秒）
AUTH_RATE_LIMIT_REQUESTS=10      # 认证接口速率限制
AUTH_RATE_LIMIT_WINDOW=60        # 认证接口速率限制窗口

# HTTPS 配置（生产环境使用）
FORCE_HTTPS=false               # 是否强制 HTTPS
TLS_CERT_FILE=                  # TLS 证书文件路径
TLS_KEY_FILE=                   # TLS 私钥文件路径
ENABLE_SECURE_HEADERS=true      # 是否启用安全头
TRUSTED_PROXIES=                # 信任的代理服务器IP
```

### 8. 请求处理配置
```bash
# === 请求配置 ===
REQUEST_TIMEOUT=30              # 请求超时时间（秒）
MAX_REQUEST_SIZE=10485760       # 最大请求大小（10MB）
ENABLE_REQUEST_ID=true          # 是否启用请求ID追踪
```

### 9. 监控和健康检查
```bash
# === 健康检查和监控 ===
HEALTH_CHECK_ENABLED=true       # 是否启用健康检查
METRICS_ENABLED=true           # 是否启用指标收集
```

### 10. 地理位置服务 (IPInfo)
```bash
# === IPInfo 地理位置服务配置 ===
IPINFO_TOKEN=5b4bddcb768db8     # IPInfo API 令牌
IPINFO_BASE_URL=https://ipinfo.io
IPINFO_TIMEOUT=10s             # API 超时时间
IPINFO_MAX_RETRIES=3           # 最大重试次数
IPINFO_RETRY_DELAY=1s          # 重试延迟
```

### 11. 第三方 API 集成
```bash
# === 第三方API集成配置 ===
# 卡片检测API配置
CARD_DETECTION_ENABLED=true
CARD_DETECTION_HOST=https://ckxiang.com
CARD_DETECTION_APP_ID=2508042205539611639
CARD_DETECTION_APP_SECRET=2caa437312d44edcaf3ab61910cf31b7
CARD_DETECTION_TIMEOUT=30
```

### 12. 管理员账户
```bash
# === 管理员默认账户（开发环境） ===
ADMIN_DEFAULT_EMAIL=admin@trusioo.com
ADMIN_DEFAULT_PASSWORD=TrusiooAdmin2024!
```

## 🚀 可选配置（按需启用）

### Redis 缓存
```bash
# REDIS_HOST=localhost
# REDIS_PORT=6379
# REDIS_PASSWORD=
# REDIS_DB=0
```

### 邮件服务
```bash
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USERNAME=your-email@gmail.com
# SMTP_PASSWORD=your-app-password
# SMTP_FROM=noreply@trusioo.com
```

### 文件上传
```bash
# UPLOAD_MAX_SIZE=5242880                              # 5MB
# UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,image/gif
# UPLOAD_PATH=./uploads
```

### 监控和追踪
```bash
# SENTRY_DSN=your-sentry-dsn
# JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

### 社交登录
```bash
# GOOGLE_CLIENT_ID=your-google-client-id
# GOOGLE_CLIENT_SECRET=your-google-client-secret
# FACEBOOK_APP_ID=your-facebook-app-id
# FACEBOOK_APP_SECRET=your-facebook-app-secret
```

### 短信服务
```bash
# SMS_PROVIDER=twilio
# TWILIO_ACCOUNT_SID=your-account-sid
# TWILIO_AUTH_TOKEN=your-auth-token
# TWILIO_FROM_PHONE=+1234567890
```

## 🛡️ 安全建议

### 生产环境配置清单
- [ ] 更改 JWT 密钥为强随机值（建议64字符以上）
- [ ] 设置 `ENV=production`
- [ ] 启用 `FORCE_HTTPS=true`
- [ ] 配置正确的 `CORS_ORIGINS`
- [ ] 设置适当的速率限制值
- [ ] 配置 TLS 证书
- [ ] 更改管理员默认密码
- [ ] 设置监控和日志收集

### 密钥生成建议
```bash
# 生成强随机 JWT 密钥
openssl rand -base64 64

# 或使用 Node.js
node -e "console.log(require('crypto').randomBytes(64).toString('base64'))"
```

## 🔧 使用方法

1. **复制并修改配置**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件填入实际值
   ```

2. **验证配置**
   ```bash
   go run cmd/main.go
   ```

3. **测试端点**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/metrics
   ```

## 📊 监控端点

- `GET /health` - 基础健康检查
- `GET /health/ready` - 就绪检查
- `GET /health/live` - 存活检查
- `GET /metrics` - 系统指标

## 🎯 性能建议

- **数据库连接池**: 根据并发量调整 `DB_MAX_OPEN_CONNS`
- **速率限制**: 根据用户量调整限制策略
- **超时设置**: 根据网络环境调整 `REQUEST_TIMEOUT`
- **日志级别**: 生产环境建议使用 `info` 或 `warn`

---

**配置完成！** 🎉 你的 Trusioo API 现在拥有完整的企业级配置，可以安全高效地运行在各种环境中。