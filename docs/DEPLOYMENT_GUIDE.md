# Trusioo API 部署指南

## 项目概述

Trusioo API 是一个完整的Go后端服务，集成了图片存储、用户认证、银行卡检测等功能。现已完成R2存储桶集成，支持高性能图片处理和CDN加速。

## 部署前准备

### 1. 系统要求

- **操作系统**: Linux (Ubuntu 20.04+ 推荐) / macOS / Windows
- **Go版本**: Go 1.24.0+
- **数据库**: PostgreSQL 12+
- **缓存**: Redis 6.0+ (可选)
- **存储**: Cloudflare R2 存储桶

### 2. 外部服务配置

#### Cloudflare R2 存储桶设置

1. **创建存储桶**：
   - 公有存储桶: `trusioo-public`
   - 私有存储桶: `trusioo-private3235`

2. **配置自定义域名**：
   - 公有桶CDN: `trusioo-public.trusioo.com`
   - 私有桶CDN: `trusioo-private.trusioo.com`

3. **获取API凭证**：
   - 访问 Cloudflare Dashboard > R2 > Manage R2 API tokens
   - 创建API令牌，记录Access Key ID和Secret Access Key

4. **存储桶权限配置**：
   ```json
   // trusioo-public 桶策略（公有读取）
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Sid": "PublicRead",
         "Effect": "Allow",
         "Principal": "*",
         "Action": "s3:GetObject",
         "Resource": "arn:aws:s3:::trusioo-public/*"
       }
     ]
   }
   ```

#### CORS 配置

在R2存储桶中配置CORS：

```json
[
  {
    "AllowedOrigins": [
      "https://trusioo.com",
      "https://admin.trusioo.com",
      "http://localhost:3000",
      "http://localhost:3001"
    ],
    "AllowedMethods": ["GET", "POST", "PUT", "DELETE"],
    "AllowedHeaders": ["*"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
```

## 环境配置

### 1. 复制配置模板

```bash
cp .env.example .env
```

### 2. 配置环境变量

编辑 `.env` 文件，填入实际配置值：

```env
# 基本配置
ENV=production
PORT=8080

# 数据库配置
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=trusioo_user
DB_PASSWORD=your-secure-password
DB_NAME=trusioo_production

# JWT 配置（生产环境必须更改）
JWT_SECRET=your-super-secure-jwt-secret-key-here
JWT_REFRESH_SECRET=your-super-secure-refresh-secret-key-here

# R2 存储配置
R2_ACCESS_KEY_ID=your-r2-access-key-id
R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
R2_ENDPOINT=https://27f7f20b92ac245bf54ced4369c47776.r2.cloudflarestorage.com
R2_REGION=auto
R2_PUBLIC_BUCKET=trusioo-public
R2_PRIVATE_BUCKET=trusioo-private3235
R2_PUBLIC_CDN_URL=https://trusioo-public.trusioo.com
R2_PRIVATE_CDN_URL=https://trusioo-private.trusioo.com

# HTTPS 配置（生产环境）
FORCE_HTTPS=true
TLS_CERT_FILE=/etc/ssl/certs/trusioo.crt
TLS_KEY_FILE=/etc/ssl/private/trusioo.key

# CORS 配置
CORS_ORIGINS=https://trusioo.com,https://admin.trusioo.com
CORS_ALLOW_ALL=false

# 安全配置
ENABLE_SECURE_HEADERS=true
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=60
```

## 部署步骤

### 1. 克隆和构建

```bash
# 克隆项目
git clone <your-repo-url>
cd trusioo_api

# 安装依赖
go mod download

# 构建项目
go build -o bin/trusioo-api cmd/main.go
```

### 2. 数据库初始化

```bash
# 创建数据库
createdb trusioo_production

# 运行迁移
make migrate-up
# 或
go run tools/db/migrate/main.go -direction=up
```

### 3. 生产环境部署

#### 使用 Systemd (推荐)

创建服务文件 `/etc/systemd/system/trusioo-api.service`：

```ini
[Unit]
Description=Trusioo API Service
After=network.target postgresql.service

[Service]
Type=simple
User=trusioo
Group=trusioo
WorkingDirectory=/opt/trusioo-api
ExecStart=/opt/trusioo-api/bin/trusioo-api
Restart=always
RestartSec=5
Environment=GIN_MODE=release

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=trusioo-api

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
# 重新加载systemd配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start trusioo-api

# 设置开机启动
sudo systemctl enable trusioo-api

# 查看状态
sudo systemctl status trusioo-api
```

#### 使用 Docker (可选)

创建 `Dockerfile`：

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/trusioo-api cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/trusioo-api .
COPY --from=builder /app/.env.example .env

CMD ["./trusioo-api"]
EXPOSE 8080
```

构建和运行：

```bash
# 构建镜像
docker build -t trusioo-api .

# 运行容器
docker run -d \
  --name trusioo-api \
  -p 8080:8080 \
  -v /path/to/.env:/root/.env \
  trusioo-api
```

### 4. Nginx 反向代理配置

创建 `/etc/nginx/sites-available/trusioo-api`：

```nginx
server {
    listen 80;
    server_name api.trusioo.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.trusioo.com;

    # SSL 配置
    ssl_certificate /etc/ssl/certs/trusioo.crt;
    ssl_certificate_key /etc/ssl/private/trusioo.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

    # 文件上传大小限制
    client_max_body_size 50M;

    # 代理到后端
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 健康检查端点
    location /health {
        access_log off;
        proxy_pass http://127.0.0.1:8080/health;
    }
}
```

启用站点：

```bash
sudo ln -s /etc/nginx/sites-available/trusioo-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 监控和维护

### 1. 日志管理

查看应用日志：

```bash
# Systemd 日志
sudo journalctl -u trusioo-api -f

# 应用日志文件
tail -f /opt/trusioo-api/logs/app.log
```

### 2. 健康检查

```bash
# 基本健康检查
curl https://api.trusioo.com/health

# 详细健康检查
curl https://api.trusioo.com/api/v1/health/detailed

# 数据库健康检查
curl https://api.trusioo.com/api/v1/health/database
```

### 3. 性能监控

设置监控指标：

```bash
# 检查API响应时间
curl -w "@curl-format.txt" -o /dev/null -s https://api.trusioo.com/health

# curl-format.txt 内容：
#     time_namelookup:  %{time_namelookup}\n
#        time_connect:  %{time_connect}\n
#     time_appconnect:  %{time_appconnect}\n
#    time_pretransfer:  %{time_pretransfer}\n
#       time_redirect:  %{time_redirect}\n
#  time_starttransfer:  %{time_starttransfer}\n
#                     ----------\n
#          time_total:  %{time_total}\n
```

### 4. 备份策略

#### 数据库备份

```bash
# 每日备份脚本
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U trusioo_user trusioo_production > /backups/trusioo_${DATE}.sql
find /backups -name "trusioo_*.sql" -mtime +7 -delete
```

#### 配置备份

```bash
# 备份关键配置文件
tar -czf /backups/config_${DATE}.tar.gz \
  /opt/trusioo-api/.env \
  /etc/nginx/sites-available/trusioo-api \
  /etc/systemd/system/trusioo-api.service
```

## 故障排除

### 1. 常见问题

**服务启动失败**：
```bash
# 检查配置文件
go run cmd/main.go -config-check

# 检查端口占用
netstat -tlnp | grep :8080
```

**数据库连接失败**：
```bash
# 测试数据库连接
psql -h DB_HOST -U DB_USER -d DB_NAME

# 检查防火墙
sudo ufw status
```

**R2存储访问失败**：
```bash
# 验证凭证
aws s3 ls --endpoint-url=https://27f7f20b92ac245bf54ced4369c47776.r2.cloudflarestorage.com

# 检查网络连通性
curl -I https://trusioo-public.trusioo.com
```

### 2. 性能优化

#### 数据库优化

```sql
-- 检查慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
WHERE mean_time > 100 
ORDER BY mean_time DESC;

-- 优化图片表索引
CREATE INDEX CONCURRENTLY idx_images_user_folder ON images(user_id, folder);
CREATE INDEX CONCURRENTLY idx_images_created_at_desc ON images(created_at DESC);
```

#### 应用优化

```env
# 调整连接池设置
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=600

# Redis缓存优化
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=10
```

## 安全配置

### 1. 防火墙规则

```bash
# UFW 配置
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw --force enable
```

### 2. SSL证书自动更新

使用Let's Encrypt：

```bash
# 安装certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d api.trusioo.com

# 自动更新
sudo crontab -e
# 添加行：0 12 * * * /usr/bin/certbot renew --quiet
```

## 扩展部署

### 1. 负载均衡

多实例部署时的Nginx配置：

```nginx
upstream trusioo_api {
    server 127.0.0.1:8080 weight=3;
    server 127.0.0.1:8081 weight=2;
    server 127.0.0.1:8082 weight=1;
}

server {
    location / {
        proxy_pass http://trusioo_api;
        # ... 其他配置
    }
}
```

### 2. 数据库集群

配置主从复制或使用云数据库服务。

### 3. CDN集成

已集成Cloudflare R2 CDN，图片自动加速全球访问。

## 联系支持

如遇到部署问题，请查看：
1. 项目文档: `/docs/`
2. 错误日志: `/logs/app.log`
3. 系统日志: `journalctl -u trusioo-api`

部署成功后，您的API将在以下地址可用：
- API基础地址: `https://api.trusioo.com`
- 健康检查: `https://api.trusioo.com/health`
- 图片上传: `https://api.trusioo.com/api/v1/images/upload`