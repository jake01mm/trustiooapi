# IPInfo Package

高性能的 IPInfo.io API 客户端，支持高并发访问和智能缓存。

## 特性

- **高并发支持**: 连接池管理，支持批量查询
- **智能缓存**: 内存缓存减少API调用
- **错误重试**: 自动重试机制
- **速率限制**: 内置速率限制防止API超限
- **上下文支持**: 完整的context.Context支持
- **配置灵活**: 支持环境变量配置

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/laitsim/trusioo/apiTrusioo/pkg/ipinfo"
)

func main() {
    // 创建配置
    config := &ipinfo.Config{
        Token:       "your-ipinfo-token",
        Timeout:     10 * time.Second,
        CacheEnable: true,
        CacheTTL:    30 * time.Minute,
    }
    
    // 创建客户端
    client := ipinfo.NewClient(config)
    defer client.Close()
    
    ctx := context.Background()
    
    // 查询单个IP
    info, err := client.GetIPInfo(ctx, "8.8.8.8")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("IP: %s, City: %s, Country: %s\n", 
        info.IP, info.City, info.Country)
}
```

### 环境变量配置

```go
// 从环境变量加载配置
config := ipinfo.LoadConfigFromEnv()
client := ipinfo.NewClient(config)
defer client.Close()
```

支持的环境变量：
- `IPINFO_TOKEN`: API token
- `IPINFO_BASE_URL`: 基础URL (默认: https://ipinfo.io)
- `IPINFO_TIMEOUT`: 请求超时时间 (如: "10s")
- `IPINFO_MAX_RETRIES`: 最大重试次数 (默认: 3)
- `IPINFO_RETRY_DELAY`: 重试延迟 (如: "1s")
- `IPINFO_CACHE_ENABLE`: 是否启用缓存 (默认: true)
- `IPINFO_CACHE_TTL`: 缓存生存时间 (如: "30m")
- `IPINFO_MAX_CONNS`: 最大连接数 (默认: 100)
- `IPINFO_MAX_IDLE_CONNS`: 最大空闲连接数 (默认: 50)

### 批量查询

```go
ips := []string{"8.8.8.8", "1.1.1.1", "208.67.222.222"}
results, err := client.BatchGetIPInfo(ctx, ips)
if err != nil {
    log.Fatal(err)
}

for ip, info := range results {
    fmt.Printf("IP: %s, City: %s\n", ip, info.City)
}
```

### 获取当前IP信息

```go
info, err := client.GetMyIP(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("My IP: %s, Location: %s\n", info.IP, info.Loc)
```

## 配置选项

```go
type Config struct {
    Token       string        // IPInfo API token
    BaseURL     string        // 基础URL
    Timeout     time.Duration // 请求超时时间
    MaxRetries  int           // 最大重试次数
    RetryDelay  time.Duration // 重试延迟
    CacheEnable bool          // 是否启用缓存
    CacheTTL    time.Duration // 缓存TTL
    MaxConns    int           // 最大连接数
    MaxIdleConns int          // 最大空闲连接数
}
```

## 性能优化

### 连接池
- 自动管理HTTP连接池
- 支持Keep-Alive连接复用
- 可配置最大连接数和空闲连接数

### 缓存机制
- 内存缓存，减少重复API调用
- 自动过期清理
- 可配置缓存TTL

### 并发控制
- 内置速率限制
- 支持批量并发查询
- 使用信号量限制并发数

## 错误处理

```go
info, err := client.GetIPInfo(ctx, "invalid-ip")
if err != nil {
    if ipErr, ok := err.(*ipinfo.Error); ok {
        switch ipErr.Code {
        case ipinfo.ErrCodeInvalidIP:
            fmt.Println("Invalid IP format")
        case ipinfo.ErrCodeRateLimit:
            fmt.Println("Rate limit exceeded")
        case ipinfo.ErrCodeUnauthorized:
            fmt.Println("Invalid token")
        default:
            fmt.Printf("Error: %s\n", ipErr.Message)
        }
    }
}
```

## 在业务模块中使用

### 在服务中集成

```go
type LocationService struct {
    ipClient ipinfo.Client
}

func NewLocationService() *LocationService {
    config := ipinfo.LoadConfigFromEnv()
    return &LocationService{
        ipClient: ipinfo.NewClient(config),
    }
}

func (s *LocationService) GetUserLocation(ctx context.Context, ip string) (*ipinfo.IPInfo, error) {
    return s.ipClient.GetIPInfo(ctx, ip)
}

func (s *LocationService) Close() error {
    return s.ipClient.Close()
}
```

### 中间件集成

```go
func IPInfoMiddleware(ipClient ipinfo.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        info, err := ipClient.GetIPInfo(c.Request.Context(), clientIP)
        if err == nil {
            c.Set("ip_info", info)
            c.Set("user_country", info.Country)
            c.Set("user_city", info.City)
        }
        
        c.Next()
    }
}
```

## 测试

```bash
go test ./pkg/ipinfo/...
```

## 许可证

MIT License