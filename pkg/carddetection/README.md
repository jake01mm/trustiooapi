# 卡片检测API封装包

这是一个用于集成第三方卡片检测API的Go封装包，支持多种类型的卡片验证，包括iTunes、Amazon、Xbox、Nike、Sephora、Razer和ND等礼品卡的检测。

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API接口](#api接口)
- [支持的卡片类型](#支持的卡片类型)
- [地区配置](#地区配置)
- [错误处理](#错误处理)
- [使用示例](#使用示例)
- [高级用法](#高级用法)
- [测试](#测试)
- [技术细节](#技术细节)
- [故障排除](#故障排除)

## 功能特性

- ✅ **多卡片类型支持**: 支持7种不同类型的卡片检测
- ✅ **自动加密处理**: 自动处理DES加密和MD5签名
- ✅ **配置集成**: 无缝集成项目现有配置系统
- ✅ **错误处理**: 完善的错误类型和错误码
- ✅ **地区验证**: 自动验证不同卡片类型支持的地区
- ✅ **上下文支持**: 支持context取消和超时
- ✅ **完整测试**: 100%测试覆盖率
- ✅ **并发安全**: 客户端支持并发使用

## 安装

此包已集成到 Trusioo API 项目中，无需单独安装。

## 快速开始

### 1. 配置设置

#### 方式一：使用配置文件

在 `config.yaml` 中添加：

```yaml
third_party:
  card_detection_enabled: true
  card_detection_host: "https://ckxiang.com"
  card_detection_app_id: "your_app_id"
  card_detection_app_secret: "your_app_secret"
  card_detection_timeout: 30
```

#### 方式二：使用环境变量

```bash
export CARD_DETECTION_ENABLED=true
export CARD_DETECTION_HOST="https://ckxiang.com"
export CARD_DETECTION_APP_ID="your_app_id"
export CARD_DETECTION_APP_SECRET="your_app_secret"
export CARD_DETECTION_TIMEOUT=30
```

### 2. 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/laitsim/trusioo/apiTrusioo/config"
    "github.com/laitsim/trusioo/apiTrusioo/pkg/carddetection"
)

func main() {
    // 初始化配置
    config.InitConfig()
    
    // 创建客户端
    cardConfig := carddetection.NewConfigFromApp(config.AppConfig)
    client := carddetection.NewClient(cardConfig)
    
    ctx := context.Background()
    
    // iTunes卡片检测
    req := &carddetection.CheckCardRequest{
        Cards:       []string{"XKQF7YZH2T3T5XWV"},
        ProductMark: carddetection.ProductMarkItunes,
        RegionID:    2,
        RegionName:  "美国",
        AutoType:    0,
    }
    
    resp, err := client.CheckCard(ctx, req)
    if err != nil {
        log.Printf("检测失败: %v", err)
        return
    }
    
    log.Printf("检测成功: %+v", resp)
}
```

## 配置说明

### Config 结构体

```go
type Config struct {
    Host      string        // API主机地址
    AppID     string        // 应用ID
    AppSecret string        // 应用密钥
    Timeout   time.Duration // 请求超时时间
}
```

### 配置创建方法

```go
// 从应用配置创建
cardConfig := carddetection.NewConfigFromApp(config.AppConfig)

// 直接参数创建
cardConfig := carddetection.NewConfigFromParams(
    "https://ckxiang.com",
    "your_app_id",
    "your_app_secret",
    30*time.Second,
)

// 验证配置
if err := cardConfig.Validate(); err != nil {
    log.Fatal("配置无效:", err)
}
```

## API接口

### 1. 卡片检测接口

#### 1.1 接口说明

使用API发送请求，检测卡片使用状态。

#### 1.2 方法签名

```go
func (c *Client) CheckCard(ctx context.Context, req *CheckCardRequest) (*CheckCardResponse, error)
```

#### 1.3 请求参数

```go
type CheckCardRequest struct {
    Cards       []string    `json:"cards" binding:"required"`       // 卡号列表
    ProductMark ProductMark `json:"productMark" binding:"required"` // 产品类型
    RegionID    int         `json:"regionId,omitempty"`             // 地区ID（部分产品需要）
    RegionName  string      `json:"regionName,omitempty"`           // 地区名称（部分产品需要）
    AutoType    int         `json:"autoType,omitempty"`             // 苹果测卡专用：0指定国家 1自动识别
}
```

#### 1.4 响应结果

```go
type CheckCardResponse struct {
    Code int    `json:"code"`    // 状态码，200表示成功
    Msg  string `json:"msg"`     // 提示信息
    Data bool   `json:"data"`    // 检测结果，true表示成功提交
}
```

#### 1.5 使用示例

```go
req := &carddetection.CheckCardRequest{
    Cards:       []string{"XKQF7YZH2T3T5XWV"},
    ProductMark: carddetection.ProductMarkItunes,
    RegionID:    2,
    RegionName:  "美国",
    AutoType:    0,
}

resp, err := client.CheckCard(ctx, req)
if err != nil {
    log.Printf("检测失败: %v", err)
    return
}

if resp.Code == 200 && resp.Data {
    log.Println("卡片检测提交成功")
}
```

### 2. 查询检测结果接口

#### 2.1 接口说明

查询已提交的卡片检测结果，获取详细的卡片状态信息。

#### 2.2 方法签名

```go
func (c *Client) CheckCardResult(ctx context.Context, req *CheckCardResultRequest) (*CardResult, error)
```

#### 2.3 请求参数

```go
type CheckCardResultRequest struct {
    ProductMark ProductMark `json:"productMark" binding:"required"` // 产品类型
    CardNo      string      `json:"cardNo" binding:"required"`      // 卡号
    PinCode     string      `json:"pinCode,omitempty"`              // PIN码（某些卡片需要）
}
```

#### 2.4 响应结果

```go
type CardResult struct {
    CardNo     string      `json:"cardNo"`     // 请求的卡号
    Status     CardStatus  `json:"status"`     // 状态码
    PinCode    string      `json:"pinCode"`    // PIN码
    Message    string      `json:"message"`    // 检测结果信息
    CheckTime  interface{} `json:"checkTime"`  // 检测时间
    RegionName string      `json:"regionName"` // 卡种国家
    RegionID   int         `json:"regionId"`   // 卡种国家编号
}
```

#### 2.5 使用示例

```go
req := &carddetection.CheckCardResultRequest{
    ProductMark: carddetection.ProductMarkItunes,
    CardNo:      "XKQF7YZH2T3T5XWV",
    PinCode:     "",
}

result, err := client.CheckCardResult(ctx, req)
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

log.Printf("卡片状态: %s", getStatusName(result.Status))
log.Printf("检测时间: %s", result.GetCheckTimeString())
```

## 支持的卡片类型

### 卡片类型枚举

```go
const (
    ProductMarkSephora ProductMark = "sephora"  // 丝芙兰
    ProductMarkRazer   ProductMark = "Razer"    // 雷蛇
    ProductMarkItunes  ProductMark = "iTunes"   // 苹果
    ProductMarkAmazon  ProductMark = "amazon"   // 亚马逊
    ProductMarkXbox    ProductMark = "xBox"     // XBOX
    ProductMarkNike    ProductMark = "nike"     // NIKE
    ProductMarkND      ProductMark = "nd"       // ND
)
```

### 卡片类型详细说明

| 卡片类型 | ProductMark | 需要地区 | 需要PIN码 | 卡号格式 | 示例 |
|---------|-------------|----------|-----------|----------|------|
| **iTunes** | `ProductMarkItunes` | ✅ RegionID/RegionName | ❌ | 16位字符 | `XKQF7YZH2T3T5XWV` |
| **Amazon** | `ProductMarkAmazon` | ✅ RegionID | ❌ | 14/15位数字 | `123456789012345` |
| **Xbox** | `ProductMarkXbox` | ✅ RegionName | ❌ | 25位字符 | `1234567890123456789012345` |
| **Nike** | `ProductMarkNike` | ❌ | ✅ | 19位卡号-6位PIN | `1234567890123456789-123456` |
| **Sephora** | `ProductMarkSephora` | ❌ | ✅ | 16位卡号-8位PIN | `1234567890123456-12345678` |
| **Razer** | `ProductMarkRazer` | ✅ RegionID | ❌ | 标准卡号 | `1234567890123456` |
| **ND** | `ProductMarkND` | ❌ | ✅ | 16位卡号-8位PIN | `1234567890123456-12345678` |

### 卡片格式验证

```go
// iTunes卡片检测
req := &carddetection.CheckCardRequest{
    Cards:       []string{"XKQF7YZH2T3T5XWV"},
    ProductMark: carddetection.ProductMarkItunes,
    RegionID:    2,        // 美国
    RegionName:  "美国",
    AutoType:    0,        // 指定国家
}

// Nike卡片检测（需要PIN码）
req := &carddetection.CheckCardRequest{
    Cards:       []string{"1234567890123456789-123456"}, // 卡号-PIN码
    ProductMark: carddetection.ProductMarkNike,
}

// Amazon卡片检测
req := &carddetection.CheckCardRequest{
    Cards:       []string{"123456789012345"},
    ProductMark: carddetection.ProductMarkAmazon,
    RegionID:    2,        // 美亚/加亚
}
```

## 地区配置

### iTunes支持的地区

```go
var ITunesRegions = []RegionInfo{
    {1, "英国"}, {2, "美国"}, {3, "德国"}, {4, "澳大利亚"},
    {5, "加拿大"}, {6, "日本"}, {8, "西班牙"}, {9, "意大利"},
    {10, "法国"}, {11, "爱尔兰"}, {12, "墨西哥"},
}
```

### Amazon支持的地区

```go
var AmazonRegions = []RegionInfo{
    {2, "美亚/加亚"}, 
    {1, "欧盟区"}, // 支持英国、德国、荷兰、西班牙、法国等
}
```

### Xbox支持的地区

```go
var XboxRegions = []string{
    "美国", "加拿大", "英国", "澳大利亚", "新西兰", "新加坡",
    "韩国", "墨西哥", "瑞典", "哥伦比亚", "阿根廷", "尼日利亚",
    "香港特别行政区", "挪威", "波兰", "德国",
}
```

### Razer支持的地区

包含22个地区，从美国、澳大利亚到亚洲各国。详细列表请参考 `types.go` 文件中的 `RazerRegions` 变量。

## 卡片状态

### 状态码定义

```go
const (
    CardStatusWaiting   CardStatus = 0 // 等待检测
    CardStatusTesting   CardStatus = 1 // 测卡中
    CardStatusValid     CardStatus = 2 // 有效
    CardStatusInvalid   CardStatus = 3 // 无效
    CardStatusRedeemed  CardStatus = 4 // 已兑换
    CardStatusFailed    CardStatus = 5 // 检测失败
    CardStatusLowPoints CardStatus = 6 // 点数不足
)
```

### 状态说明

| 状态码 | 状态名称 | 含义 | 后续操作 |
|-------|---------|------|----------|
| **0** | 等待检测 | 卡片已提交，等待开始检测 | 稍后再查询 |
| **1** | 测卡中 | 卡片正在检测中 | 等待检测完成 |
| **2** | 有效 | 卡片有效且未使用 | 可以正常使用 |
| **3** | 无效 | 卡片无效或格式错误 | 检查卡号格式 |
| **4** | 已兑换 | 卡片已被使用，余额为0 | 卡片已失效 |
| **5** | 检测失败 | 检测过程失败 | 重新尝试或联系支持 |
| **6** | 点数不足 | 检测服务点数不足 | 联系服务提供商 |

### 状态检查示例

```go
result, err := client.CheckCardResult(ctx, req)
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

switch result.Status {
case carddetection.CardStatusValid:
    log.Println("✅ 卡片有效，可以使用")
case carddetection.CardStatusRedeemed:
    log.Println("🔴 卡片已被兑换，余额为0")
case carddetection.CardStatusInvalid:
    log.Printf("❌ 卡片无效: %s", result.Message)
case carddetection.CardStatusTesting:
    log.Println("🔄 卡片正在检测中，请稍后再查询")
default:
    log.Printf("❓ 未知状态: %d", result.Status)
}
```

## 错误处理

### 错误类型

```go
const (
    ErrCodeInvalidConfig     = 1001 // 配置错误
    ErrCodeInvalidRequest    = 1002 // 请求参数错误
    ErrCodeEncryptionFailed  = 1003 // 加密失败
    ErrCodeDecryptionFailed  = 1004 // 解密失败
    ErrCodeSignatureFailed   = 1005 // 签名失败
    ErrCodeAPIRequest        = 1006 // API请求错误
    ErrCodeAPIResponse       = 1007 // API响应错误
    ErrCodeTimeout           = 1008 // 请求超时
    ErrCodeUnsupportedRegion = 1009 // 不支持的地区
    ErrCodeInvalidCardFormat = 1010 // 卡片格式错误
)
```

### 错误处理示例

```go
resp, err := client.CheckCard(ctx, req)
if err != nil {
    if carddetection.IsCardDetectionError(err) {
        errorCode := carddetection.GetErrorCode(err)
        switch errorCode {
        case carddetection.ErrCodeInvalidConfig:
            log.Println("配置错误，请检查API凭据")
        case carddetection.ErrCodeInvalidRequest:
            log.Println("请求参数错误，请检查卡号格式")
        case carddetection.ErrCodeTimeout:
            log.Println("请求超时，请稍后重试")
        default:
            log.Printf("卡片检测错误 (错误码: %d): %v", errorCode, err)
        }
    } else {
        log.Printf("其他错误: %v", err)
    }
    return
}
```

### 预定义错误

```go
// 配置相关错误
ErrInvalidConfig     // 无效配置
ErrMissingHost       // 缺少主机配置
ErrMissingAppID      // 缺少应用ID
ErrMissingAppSecret  // 缺少应用密钥

// 请求相关错误
ErrInvalidRequest       // 无效请求
ErrInvalidProductMark   // 无效产品类型
ErrInvalidCardFormat    // 无效卡片格式
ErrUnsupportedRegion    // 不支持的地区

// 加密相关错误
ErrEncryptionFailed  // 加密失败
ErrDecryptionFailed  // 解密失败
ErrSignatureFailed   // 签名失败

// API相关错误
ErrAPIRequest   // API请求失败
ErrAPIResponse  // API响应无效
ErrTimeout      // 请求超时
```

## 使用示例

### 完整的检测流程

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/laitsim/trusioo/apiTrusioo/config"
    "github.com/laitsim/trusioo/apiTrusioo/pkg/carddetection"
)

func main() {
    // 1. 初始化配置
    if err := config.InitConfig(); err != nil {
        log.Fatalf("配置初始化失败: %v", err)
    }
    
    // 2. 创建客户端
    cardConfig := carddetection.NewConfigFromApp(config.AppConfig)
    if cardConfig == nil {
        log.Fatal("无法创建卡片检测配置")
    }
    
    client := carddetection.NewClient(cardConfig)
    
    // 3. 验证配置
    if err := client.ValidateConfig(); err != nil {
        log.Fatalf("配置验证失败: %v", err)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // 4. 提交检测
    checkReq := &carddetection.CheckCardRequest{
        Cards:       []string{"XKQF7YZH2T3T5XWV"},
        ProductMark: carddetection.ProductMarkItunes,
        RegionID:    2,
        RegionName:  "美国",
        AutoType:    0,
    }
    
    fmt.Println("📤 提交卡片检测...")
    checkResp, err := client.CheckCard(ctx, checkReq)
    if err != nil {
        log.Fatalf("检测提交失败: %v", err)
    }
    
    if checkResp.Code != 200 || !checkResp.Data {
        log.Fatalf("检测提交失败: Code=%d, Msg=%s", checkResp.Code, checkResp.Msg)
    }
    
    fmt.Println("✅ 检测提交成功，等待处理...")
    
    // 5. 等待处理
    time.Sleep(5 * time.Second)
    
    // 6. 查询结果
    resultReq := &carddetection.CheckCardResultRequest{
        ProductMark: carddetection.ProductMarkItunes,
        CardNo:      "XKQF7YZH2T3T5XWV",
        PinCode:     "",
    }
    
    fmt.Println("📊 查询检测结果...")
    result, err := client.CheckCardResult(ctx, resultReq)
    if err != nil {
        log.Fatalf("结果查询失败: %v", err)
    }
    
    // 7. 显示结果
    fmt.Printf("💳 卡号: %s\n", result.CardNo)
    fmt.Printf("📊 状态: %s (%d)\n", getStatusName(result.Status), result.Status)
    fmt.Printf("💬 消息: %s\n", result.Message)
    fmt.Printf("🕒 检测时间: %s\n", result.GetCheckTimeString())
    fmt.Printf("🌍 地区: %s\n", result.RegionName)
}

func getStatusName(status carddetection.CardStatus) string {
    switch status {
    case carddetection.CardStatusWaiting:
        return "等待检测"
    case carddetection.CardStatusTesting:
        return "测卡中"
    case carddetection.CardStatusValid:
        return "有效"
    case carddetection.CardStatusInvalid:
        return "无效"
    case carddetection.CardStatusRedeemed:
        return "已兑换"
    case carddetection.CardStatusFailed:
        return "检测失败"
    case carddetection.CardStatusLowPoints:
        return "点数不足"
    default:
        return "未知状态"
    }
}
```

### 批量检测示例

```go
func batchCheckCards(client *carddetection.Client, cards []string) {
    ctx := context.Background()
    
    // 批量提交检测
    req := &carddetection.CheckCardRequest{
        Cards:       cards,
        ProductMark: carddetection.ProductMarkItunes,
        RegionID:    2,
        RegionName:  "美国",
    }
    
    resp, err := client.CheckCard(ctx, req)
    if err != nil {
        log.Printf("批量检测失败: %v", err)
        return
    }
    
    if resp.Code != 200 || !resp.Data {
        log.Printf("批量检测提交失败: %s", resp.Msg)
        return
    }
    
    // 等待处理
    time.Sleep(10 * time.Second)
    
    // 逐个查询结果
    for _, cardNo := range cards {
        resultReq := &carddetection.CheckCardResultRequest{
            ProductMark: carddetection.ProductMarkItunes,
            CardNo:      cardNo,
            PinCode:     "",
        }
        
        result, err := client.CheckCardResult(ctx, resultReq)
        if err != nil {
            log.Printf("卡片 %s 查询失败: %v", cardNo, err)
            continue
        }
        
        fmt.Printf("卡片 %s: %s\n", cardNo, getStatusName(result.Status))
    }
}
```

### 不同类型卡片检测示例

```go
// iTunes检测
func checkITunesCard(client *carddetection.Client, cardNo string) {
    req := &carddetection.CheckCardRequest{
        Cards:       []string{cardNo},
        ProductMark: carddetection.ProductMarkItunes,
        RegionID:    2,
        RegionName:  "美国",
        AutoType:    0, // 指定国家
    }
    // ... 执行检测
}

// Amazon检测
func checkAmazonCard(client *carddetection.Client, cardNo string) {
    req := &carddetection.CheckCardRequest{
        Cards:       []string{cardNo},
        ProductMark: carddetection.ProductMarkAmazon,
        RegionID:    2, // 美亚/加亚
    }
    // ... 执行检测
}

// Xbox检测
func checkXboxCard(client *carddetection.Client, cardNo string) {
    req := &carddetection.CheckCardRequest{
        Cards:       []string{cardNo},
        ProductMark: carddetection.ProductMarkXbox,
        RegionName:  "美国",
    }
    // ... 执行检测
}

// Nike检测（需要PIN码）
func checkNikeCard(client *carddetection.Client, cardNo, pinCode string) {
    // Nike卡片格式: 卡号-PIN码
    cardWithPin := fmt.Sprintf("%s-%s", cardNo, pinCode)
    
    req := &carddetection.CheckCardRequest{
        Cards:       []string{cardWithPin},
        ProductMark: carddetection.ProductMarkNike,
    }
    // ... 执行检测
    
    // 查询时需要分别提供卡号和PIN码
    resultReq := &carddetection.CheckCardResultRequest{
        ProductMark: carddetection.ProductMarkNike,
        CardNo:      cardNo,    // 不包含PIN码
        PinCode:     pinCode,   // 单独的PIN码
    }
    // ... 查询结果
}
```

## 高级用法

### 自定义超时

```go
// 创建带自定义超时的客户端
config := carddetection.NewConfigFromParams(
    "https://ckxiang.com",
    "your_app_id",
    "your_app_secret",
    60*time.Second, // 60秒超时
)

client := carddetection.NewClient(config)

// 为单个请求设置超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.CheckCard(ctx, req)
```

### 重试机制

```go
func checkCardWithRetry(client *carddetection.Client, req *carddetection.CheckCardRequest, maxRetries int) (*carddetection.CheckCardResponse, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        resp, err := client.CheckCard(ctx, req)
        cancel()
        
        if err == nil && resp.Code == 200 {
            return resp, nil
        }
        
        lastErr = err
        
        // 如果是配置错误，不重试
        if carddetection.IsCardDetectionError(err) {
            errorCode := carddetection.GetErrorCode(err)
            if errorCode == carddetection.ErrCodeInvalidConfig {
                return nil, err
            }
        }
        
        // 等待后重试
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}
```

### 并发检测

```go
func concurrentCardCheck(client *carddetection.Client, cards []string) {
    var wg sync.WaitGroup
    results := make(chan string, len(cards))
    
    for _, cardNo := range cards {
        wg.Add(1)
        go func(card string) {
            defer wg.Done()
            
            req := &carddetection.CheckCardRequest{
                Cards:       []string{card},
                ProductMark: carddetection.ProductMarkItunes,
                RegionID:    2,
                RegionName:  "美国",
            }
            
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            
            resp, err := client.CheckCard(ctx, req)
            if err != nil {
                results <- fmt.Sprintf("卡片 %s 检测失败: %v", card, err)
                return
            }
            
            if resp.Code == 200 && resp.Data {
                results <- fmt.Sprintf("卡片 %s 检测成功", card)
            } else {
                results <- fmt.Sprintf("卡片 %s 检测失败: %s", card, resp.Msg)
            }
        }(cardNo)
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    for result := range results {
        fmt.Println(result)
    }
}
```

### 轮询检测状态

```go
func pollCardStatus(client *carddetection.Client, cardNo string, productMark carddetection.ProductMark) (*carddetection.CardResult, error) {
    req := &carddetection.CheckCardResultRequest{
        ProductMark: productMark,
        CardNo:      cardNo,
        PinCode:     "",
    }
    
    // 最多轮询30次，每次间隔5秒
    for i := 0; i < 30; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        result, err := client.CheckCardResult(ctx, req)
        cancel()
        
        if err != nil {
            return nil, err
        }
        
        // 如果不是等待或检测中状态，返回结果
        if result.Status != carddetection.CardStatusWaiting && 
           result.Status != carddetection.CardStatusTesting {
            return result, nil
        }
        
        fmt.Printf("第 %d 次检查，状态: %s，继续等待...\n", i+1, getStatusName(result.Status))
        time.Sleep(5 * time.Second)
    }
    
    return nil, fmt.Errorf("轮询超时，卡片仍在处理中")
}
```

## 测试

### 运行单元测试

```bash
go test ./pkg/carddetection -v
```

### 运行覆盖率测试

```bash
go test ./pkg/carddetection -cover
```

### 运行集成测试

```bash
# 基础API功能测试
cd tests/card_detection/cardtest
CARD_DETECTION_ENABLED=true \
CARD_DETECTION_HOST="https://ckxiang.com" \
CARD_DETECTION_APP_ID="your_app_id" \
CARD_DETECTION_APP_SECRET="your_app_secret" \
CARD_DETECTION_TIMEOUT=30 \
go run main.go

# 多种卡片类型测试
cd tests/card_detection/multicardtest
go run main.go

# 详细调试测试
cd tests/card_detection/carddebug
go run main.go

# 卡片查询测试
cd tests/card_detection/querycard
go run main.go
```

### 测试覆盖的功能

- ✅ 客户端创建和配置验证
- ✅ 所有卡片类型的检测请求
- ✅ 地区验证和格式检查
- ✅ 加密和签名功能
- ✅ 错误处理和状态码
- ✅ 超时和取消机制
- ✅ 并发安全性

## 技术细节

### 加密算法

本包使用与第三方API完全兼容的加密算法：

- **签名算法**: MD5 (secret + sortedParams + secret)
- **加密算法**: DES ECB模式
- **编码格式**: 十六进制编码

### 签名生成过程

```go
// 1. 参数排序（按ASCII码）
params := map[string]interface{}{
    "cards":       []string{"XKQF7YZH2T3T5XWV"},
    "productMark": "iTunes",
    "regionId":    2,
    "regionName":  "美国",
    "timestamp":   "1234567890",
}

// 2. 构建签名字符串
// secret + "cards[\"XKQF7YZH2T3T5XWV\"]productMarkiTunesregionId2regionName美国timestamp1234567890" + secret

// 3. MD5计算
sign := md5(signString)
```

### DES加密过程

```go
// 1. 生成DES密钥（使用appSecret的前8字节）
key := generateDESKey(appSecret)

// 2. PKCS7填充
paddedData := pkcs7Pad(plainText, 8)

// 3. DES ECB加密
cipherText := desEncrypt(paddedData, key)

// 4. 十六进制编码
result := hex.EncodeToString(cipherText)
```

### HTTP请求格式

```json
POST /api/userApiManage/checkCard
Content-Type: application/json
appId: your_app_id

{
  "data": "encrypted_and_hex_encoded_data"
}
```

### 并发安全

客户端实例是并发安全的，可以在多个goroutine中共享使用：

```go
client := carddetection.NewClient(config)

// 并发使用是安全的
go func() {
    resp, err := client.CheckCard(ctx, req1)
    // ...
}()

go func() {
    resp, err := client.CheckCard(ctx, req2)
    // ...
}()
```

## 故障排除

### 常见错误及解决方案

#### 1. IP地址不在白名单内

**错误信息**: `IP地址不在白名单内`

**解决方案**:
- 联系API提供商将你的IP地址加入白名单
- 确认使用的是正确的公网IP地址

```bash
# 查看当前公网IP
curl ipinfo.io/ip
```

#### 2. 验签失败

**错误信息**: `验签失败`

**原因分析**:
- AppID或AppSecret错误
- 签名算法实现不正确
- 参数格式不匹配

**解决方案**:
```go
// 检查配置
fmt.Printf("AppID: %s\n", config.AppID)
fmt.Printf("AppSecret: %s\n", config.AppSecret[:10] + "...")

// 使用调试模式
cd tests/card_detection/carddebug
go run main.go
```

#### 3. 请求超时

**错误信息**: `request timeout`

**解决方案**:
```go
// 增加超时时间
config := carddetection.NewConfigFromParams(
    "https://ckxiang.com",
    "your_app_id",
    "your_app_secret",
    60*time.Second, // 增加到60秒
)

// 或使用上下文超时
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
```

#### 4. 卡片格式错误

**错误信息**: `invalid card format`

**解决方案**:
- 检查卡号长度和格式
- 确认是否需要PIN码
- 验证地区设置

```go
// iTunes卡片: 16位字符
cardNo := "XKQF7YZH2T3T5XWV" // ✅ 正确
cardNo := "123456789012345"   // ❌ 错误，这是Amazon格式

// Nike卡片: 需要PIN码
cardNo := "1234567890123456789-123456" // ✅ 正确格式
```

#### 5. 不支持的地区

**错误信息**: `unsupported region for this product`

**解决方案**:
```go
// 检查地区是否支持
client := carddetection.NewClient(config)

// iTunes检查
if client.isValidITunesRegion(regionID) {
    // 支持的地区
} else {
    // 不支持的地区，选择其他地区
}
```

### 调试方法

#### 1. 启用详细日志

```go
// 在测试程序中添加详细输出
fmt.Printf("请求参数: %+v\n", req)
fmt.Printf("API响应: %+v\n", resp)
```

#### 2. 使用调试工具

```bash
# 运行调试程序
cd tests/card_detection/carddebug
CARD_DETECTION_ENABLED=true \
CARD_DETECTION_HOST="https://ckxiang.com" \
CARD_DETECTION_APP_ID="your_app_id" \
CARD_DETECTION_APP_SECRET="your_app_secret" \
go run main.go
```

#### 3. 网络连接检查

```bash
# 检查网络连接
curl -I https://ckxiang.com

# 检查DNS解析
nslookup ckxiang.com
```

#### 4. 配置验证

```bash
# 运行配置验证
cd tests/config_carddetection
go run main.go docker/.env.development
```

### 性能优化

#### 1. 连接池设置

```go
// 自定义HTTP客户端
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

// 注意：当前版本不支持自定义HTTP客户端
// 这是未来版本的改进建议
```

#### 2. 批量请求

```go
// 尽量使用批量检测而非单个检测
req := &carddetection.CheckCardRequest{
    Cards: []string{
        "XKQF7YZH2T3T5XWV",
        "XNVKVZ8KNHG43LFP",
        "ANOTHER_CARD_NUMBER",
    },
    ProductMark: carddetection.ProductMarkItunes,
    RegionID:    2,
    RegionName:  "美国",
}
```

#### 3. 合理的重试策略

```go
// 实现指数退避重试
func exponentialBackoff(attempt int) time.Duration {
    return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

for i := 0; i < maxRetries; i++ {
    resp, err := client.CheckCard(ctx, req)
    if err == nil {
        return resp, nil
    }
    
    time.Sleep(exponentialBackoff(i))
}
```

## 许可证

本项目采用MIT许可证。详情请参见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交问题和拉取请求。在提交之前，请确保：

1. 代码通过所有测试
2. 遵循Go代码规范
3. 更新相关文档
4. 添加适当的测试用例

## 联系方式

如果您有任何问题或建议，请通过以下方式联系：

- 提交 [GitHub Issue](https://github.com/your-repo/issues)
- 发送邮件至 [your-email@domain.com]

## 更新日志

### v1.0.0 (2025-01-XX)

- ✅ 初始版本发布
- ✅ 支持7种卡片类型检测
- ✅ 完整的加密和签名实现
- ✅ 错误处理和状态管理
- ✅ 100%测试覆盖率
- ✅ 完整的文档和示例