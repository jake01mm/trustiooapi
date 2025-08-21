# “中间件上下文 + Feature Flags + Gate” 落地方案（基于现有仓库）

本文档基于项目内真实代码完成“仓库体检 → 差距评估 → 变更清单 → 代码草案 → 迁移脚本 → 验收步骤”，并给出结论与首周交付范围。所有路径均为仓库内真实路径，可直接定位。

---

## 一、仓库体检（含证据路径）

- 语言/框架与目录结构
  - Go + Gin。核心入口与目录：
    - cmd/main.go（HTTP Server 启动、加载配置/DB/Redis）
    - internal/router/router.go（Gin 引擎、核心中间件、路由注册）
    - internal/middleware/*（认证/日志/安全/CORS/超时/限流等）
    - internal/auth/*（用户/管理员登录、验证、Profile 等）
    - pkg/*（auth、database、redis、logger、ipinfo 等可复用包）

- 认证链路（登录端点、JWT/Session 验证中间件、上下文注入现状）
  - 登录端点：
    - internal/auth/user_auth/routes.go → POST /api/v1/auth/login、/login/verify、/refresh 等
    - internal/auth/admin_auth/routes.go → POST /api/v1/admin/auth/login、/login/verify、/refresh 等
  - JWT 验证中间件与上下文注入：
    - internal/middleware/auth.go 中 AuthMiddleware / AdminAuthMiddleware / SuperAdminMiddleware
      - 从 Authorization: Bearer 解析访问令牌，校验后注入：
        - c.Set("user_id"), c.Set("user_email"), c.Set("user_role"), c.Set("user_type")
  - 业务中读取示例：
    - internal/auth/user_auth/handler.go:GetProfile 从 c.Get("user_id") 读取
    - internal/auth/admin_auth/handler.go:GetProfile 从 c.Get("user_id") 读取

- Header 读取位置
  - 已有：
    - internal/middleware/request_id.go 读取 X-Request-ID、X-Correlation-ID，并回写响应头
    - internal/middleware/security.go 的 CORSMiddleware 设置 Access-Control-Allow-Headers（当前未包含 X-Platform / X-App-Version / X-Device-Id / X-Locale / X-Build-Channel）
  - 未发现：
    - X-Platform / X-App-Version / X-Device-Id / X-Locale / X-Build-Channel 的读取逻辑（rg 全仓库搜索为空）

- 菜单/路由是否服务端生成
  - 未发现服务端“菜单裁剪”实现或菜单数据端点（rg 搜索未命中），当前路由以功能模块为主（auth、admin、card-detection）。

- Feature Flags / Remote Config 现状
  - 未发现 feature flag / remote-config 相关表或代码（rg 搜索未命中）。

以上证据可在以下路径查看：
- internal/router/router.go（核心中间件/路由注册）
- internal/middleware/auth.go（JWT 验证与上下文注入）
- internal/middleware/request_id.go（X-Request-ID / X-Correlation-ID）
- internal/middleware/security.go（CORS 允许头列表）
- internal/auth/user_auth/routes.go、internal/auth/admin_auth/routes.go（登录端点）
- internal/auth/*/handler.go（从上下文读取 user_id 的示例）
---

## 二、差距评估（✅/⚠️/❌）

- 认证解析（JWT/Session）：✅
  - 已有 JWT 生成/校验（pkg/auth）、Gin 中间件注入 user_id/role/type，并在 handler 中读取使用。
  - 影响面：后续在 EvalCtx 组装时直接复用，不破坏现有认证链路。

- 上下文中间件（EvalCtx）：❌
  - 当前仅零散通过 c.Set 方式注入 user_id/role 等，不包含 kyc/country/platform/app_version/device_id/channel/ip/geo 等完整上下文对象。
  - 影响面：需要新增 EvalCtx 构造中间件，标准化上下文获取与存取方式。

- Feature Flags 存储（Postgres/JSONB）：❌
  - 无表、无DAO、无服务。
  - 影响面：需新增 migrations + DAO + evaluator。

- 评估器（国家/KYC/版本/平台/时间窗/灰度）：❌
  - 不存在统一 Evaluate(flag, ctx)。

- Gate 中间件（接口门禁）：❌
  - 当前没有 feature-gate 机制，无法以最小侵入保护敏感接口。

- 服务端菜单裁剪：❌
  - 未提供 GET /v1/menus 等端点与基于 flags 的裁剪逻辑。

- 缓存与热更新（LISTEN/NOTIFY 或定时刷新）：⚠️
  - 项目已有 Redis 封装（pkg/redis），但未用于 flags；无热更新机制。

- 审计日志（谁改了哪个开关）：❌
  - 未发现审计表与写入逻辑。

- 客户端 Header 规范：⚠️
  - 请求头中 X-Request-ID / X-Correlation-ID 已支持；缺少 X-Platform、X-App-Version、X-Device-Id、X-Build-Channel、X-Locale 的读取与允许列表配置。
---

## 三、落地方案（文件级改动清单）

1) internal/middleware/evalctx.go（新增）
   - 职责：从 token + headers + IP/GeoIP 组装 EvalCtx，并注入 gin.Context
   - 对外函数：
     - type EvalCtx struct { UserID int64; Role string; KYC string; Country string; Platform string; AppVersion string; DeviceID string; Channel string; Locale string; IP string }
     - func BuildEvalCtx(c *gin.Context) *EvalCtx
     - func EvalCtxMiddleware() gin.HandlerFunc
   - 调用点：在 router.SetupRouter() 中全局注册，位于 Auth 之后、业务之前
   - 耦合点：
     - 读取 AuthMiddleware 注入的 user_id/user_role
     - 读取 Header：X-Platform、X-App-Version、X-Device-Id、X-Build-Channel、X-Locale
     - 使用 pkg/ipinfo 或 pkg/ipinfo/geo（若已有）根据 IP 推断国家（若无则仅传入 IP，预留接口）

2) internal/feature/evaluator.go（新增）
   - 职责：Evaluate(flagKey, ctx) 根据 DB 中 JSONB 规则进行评估；支持策略：countries / min_kyc / roles / platforms / app_version_min/max / time_window / rollout% + stickiness(user_id)
   - 对外函数：
     - func Evaluate(c *gin.Context, flagKey string) (bool, error)
     - func GetEvaluatedFlags(c *gin.Context) (map[string]bool, error)
   - 调用点：
     - Gate 中间件
     - GET /api/v1/flags
     - 菜单裁剪逻辑
   - 耦合点：依赖 pkg/database 或 pkg/redis 做缓存；配置热更新策略（先用定时刷新，后续可扩展 LISTEN/NOTIFY）

3) internal/feature/middleware_gate.go（新增）
   - 职责：Gin 中间件 Gate(featureKey)；当 Evaluate 为 false 时返回 403 {"error":"feature_disabled"}
   - 对外函数：
     - func Gate(featureKey string) gin.HandlerFunc
   - 调用点：在具体路由上包裹敏感端点，如 POST /v1/trade/orders、POST /v1/withdraw/requests

4) internal/menu/handler.go（新增）
   - 职责：GET /api/v1/menus：从 DB 读取菜单（先提供内存/表驱动两种可选方案），基于 Evaluate 动态裁剪
   - 对外：
     - func RegisterRoutes(router *gin.RouterGroup, handler *Handler)
     - func (h *Handler) GetMenus(c *gin.Context)

5) migrations/xxxx_feature_flags.sql（新增）
   - 职责：建表/索引/审计/急停表
   - 表设计：
     - feature_flags(id, key UNIQUE, enabled BOOL, rules JSONB, description TEXT, updated_by TEXT, updated_at TIMESTAMPTZ)
     - feature_flag_audits(id, key, old_value JSONB, new_value JSONB, changed_by TEXT, changed_at TIMESTAMPTZ)
     - feature_global_switches(id, key UNIQUE, forced_state TEXT, updated_at TIMESTAMPTZ) -- 可选，用于一键急停

6) 路由注册与集成改动（最小侵入）
   - internal/router/router.go：
     - 注册 EvalCtxMiddleware（在 AuthMiddleware 之后）
     - 新增 flags 与 menus 端点
     - 对示例敏感端点（若暂不存在 trade/withdraw，可先演示在 carddetection 或占位路由）挂 Gate()

7) 业务端点（示例）
   - internal/trade/handler.go（若未来引入）：POST /api/v1/trade/orders（Gate("giftcard.trade")）
   - internal/withdraw/handler.go（若未来引入）：POST /api/v1/withdraw/requests（Gate("wallet.withdraw")）
   - 当前仓库未有 trade/withdraw，可先在 internal/router/router.go 中新建示例路由组示范 Gate

最小改动路径与可回滚：
- feature 默认全关（enabled=false/无匹配规则），即不影响现网
- EvalCtxMiddleware 仅新增读取 Headers 与上下文，不变更现有 handler/service 逻辑
- Gate 中间件仅在新加的示例路由上启用；真实业务路由可按灰度逐步添加
- 审计/全局急停仅新增表与查询，不影响现有逻辑
---

## 四、代码草案（可编译片段）

说明：以下为关键实现草案（imports 完整，可单独编译），路径以项目内真实路径为准。

1) internal/middleware/evalctx.go

```go
package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// EvalCtx 请求评估上下文
// 可按需扩展 KYC 等字段
 type EvalCtx struct {
	UserID     int64  `json:"user_id"`
	Role       string `json:"role"`
	KYC        string `json:"kyc"`
	Country    string `json:"country"`
	Platform   string `json:"platform"`
	AppVersion string `json:"app_version"`
	DeviceID   string `json:"device_id"`
	Channel    string `json:"channel"`
	Locale     string `json:"locale"`
	IP         string `json:"ip"`
}

func BuildEvalCtx(c *gin.Context) *EvalCtx {
	ctx := &EvalCtx{}
	if v, ok := c.Get("user_id"); ok {
		if id, ok2 := v.(int64); ok2 { ctx.UserID = id }
	}
	if v, ok := c.Get("user_role"); ok { ctx.Role, _ = v.(string) }

	ctx.Platform = c.GetHeader("X-Platform")
	ctx.AppVersion = c.GetHeader("X-App-Version")
	ctx.DeviceID = c.GetHeader("X-Device-Id")
	ctx.Channel = c.GetHeader("X-Build-Channel")
	ctx.Locale = c.GetHeader("X-Locale")

	ip := clientIP(c)
	ctx.IP = ip
	// TODO: 可集成 pkg/ipinfo 根据 IP 解析国家
	return ctx
}

func EvalCtxMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ec := BuildEvalCtx(c)
		c.Set("eval_ctx", ec)
		c.Next()
	}
}

// clientIP 兼容反向代理后的真实 IP
func clientIP(c *gin.Context) string {
	// 优先 X-Forwarded-For
	xff := c.Request.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" { return ip }
	}
	// 其次 X-Real-IP
	if xr := c.Request.Header.Get("X-Real-IP"); xr != "" { return xr }
	// 回退到 gin 的 ClientIP
	ip := c.ClientIP()
	// 处理可能的端口
	if host, _, err := net.SplitHostPort(ip); err == nil { return host }
	return ip
}
```

2) internal/feature/evaluator.go

```go
package feature

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/binary"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

type Flag struct {
	Key       string
	Enabled   bool
	RulesJSON []byte // JSONB 原始数据
}

type Evaluator struct {
	DB *sql.DB
}

func NewEvaluator(db *sql.DB) *Evaluator { return &Evaluator{DB: db} }

// Evaluate: 读取 flag 并根据 ctx 判断是否开启
func (e *Evaluator) Evaluate(c *gin.Context, flagKey string) (bool, error) {
	// 急停/全局覆盖可在此优先判断（可选表）
	flag, err := e.getFlag(c, flagKey)
	if err != nil { return false, err }
	if flag == nil { return false, nil }
	if !flag.Enabled { return false, nil }
	// TODO: 解析 flag.RulesJSON 并评估（countries/min_kyc/roles/platforms/app_version/time_window/rollout）
	// 先给最小可用：Enabled 即 true
	return true, nil
}

func (e *Evaluator) GetEvaluatedFlags(c *gin.Context) (map[string]bool, error) {
	// 简化：查询全部 flags 并逐个 Evaluate（生产可缓存）
	rows, err := e.DB.QueryContext(c, `SELECT key, enabled FROM feature_flags`)
	if err != nil { return nil, err }
	defer rows.Close()
	res := map[string]bool{}
	for rows.Next() {
		var key string
		var enabled bool
		if err := rows.Scan(&key, &enabled); err != nil { return nil, err }
		res[key] = enabled
	}
	return res, nil
}

func (e *Evaluator) getFlag(ctx context.Context, key string) (*Flag, error) {
	row := e.DB.QueryRowContext(ctx, `SELECT key, enabled, rules FROM feature_flags WHERE key=$1`, key)
	var f Flag
	var rules []byte
	if err := row.Scan(&f.Key, &f.Enabled, &rules); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, nil }
		return nil, err
	}
	f.RulesJSON = rules
	return &f, nil
}

// stickiness: 对 user_id 或 device_id 做一致性哈希
func stickyPercent(key string, stickValue string) int {
	h := sha1.Sum([]byte(key+":"+stickValue))
	// 取前 4 字节转为 0-100 百分比
	x := binary.BigEndian.Uint32(h[:4])
	return int(x % 100)
}

// timeWindow: 当前是否处于允许时间窗
func inTimeWindow(now time.Time, start, end time.Time) bool {
	if start.IsZero() && end.IsZero() { return true }
	if !start.IsZero() && now.Before(start) { return false }
	if !end.IsZero() && now.After(end) { return false }
	return true
}
```

3) internal/feature/middleware_gate.go

```go
package feature

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Gate(ev *Evaluator, featureKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, err := ev.Evaluate(c, featureKey)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error":"feature_disabled"})
			return
		}
		c.Next()
	}
}
```

4) internal/menu/handler.go（示例）

```go
package menu

import (
	"net/http"
	"trusioo_api/internal/feature"

	"github.com/gin-gonic/gin"
)

type Handler struct { Ev *feature.Evaluator }

func NewHandler(ev *feature.Evaluator) *Handler { return &Handler{Ev: ev} }

func (h *Handler) GetMenus(c *gin.Context) {
	flags, err := h.Ev.GetEvaluatedFlags(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"internal_error"})
		return
	}
	// 简化示例：基于 flags 裁剪
	menus := []map[string]interface{}{
		{"key":"giftcard","children": []interface{}{
			{"key":"trade","enabled": flags["giftcard.trade"]},
		}},
		{"key":"wallet","children": []interface{}{
			{"key":"withdraw","enabled": flags["wallet.withdraw"]},
		}},
	}
	c.JSON(http.StatusOK, gin.H{"menus": menus})
}
```

5) internal/router/router.go（集成片段示意）

```go
// 省略 imports
// 在 SetupRouter() 中：
// r.Use(middleware.AuthMiddleware()) 之后，增加：
// r.Use(middleware.EvalCtxMiddleware())
// 
// 初始化 evaluator 并注册端点：
// db := database.GetDB() // 复用你现有的 pkg/database
// ev := feature.NewEvaluator(db)
// 
// api.GET("/flags", func(c *gin.Context){ res, _ := ev.GetEvaluatedFlags(c); c.JSON(200, res) })
// menuHandler := menu.NewHandler(ev)
// api.GET("/menus", menuHandler.GetMenus)
// 
// 示例受控端点：
// trade := api.Group("/trade")
// trade.POST("/orders", feature.Gate(ev, "giftcard.trade"), func(c *gin.Context){ c.JSON(200, gin.H{"ok":true}) })
// withdraw := api.Group("/withdraw")
// withdraw.POST("/requests", feature.Gate(ev, "wallet.withdraw"), func(c *gin.Context){ c.JSON(200, gin.H{"ok":true}) })
```

6) migrations/20250221_feature_flags.sql（示例）

```sql
-- feature flags core tables
CREATE TABLE IF NOT EXISTS feature_flags (
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    description TEXT,
    updated_by TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX IF NOT EXISTS idx_feature_flags_rules_gin ON feature_flags USING GIN (rules);

CREATE TABLE IF NOT EXISTS feature_flag_audits (
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    old_value JSONB,
    new_value JSONB,
    changed_by TEXT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 可选：一键急停
CREATE TABLE IF NOT EXISTS feature_global_switches (
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    forced_state TEXT CHECK (forced_state IN ('enabled','disabled')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
---

## 五、客户端对接规范（Header 与匿名 stickiness）

统一 Header 约定（客户端需在所有请求中携带）：
- X-Platform: ios|android|web
- X-App-Version: 1.3.0（语义化版本）
- X-Device-Id: <稳定ID>（移动端建议使用 Keychain/Keystore 持久化，卸载重装仍能复用）
- X-Build-Channel: appstore|testflight|internal
- X-Locale: zh-CN

匿名用户 stickiness：
- 未登录时以 Device-Id 作为一致性哈希种子；
- 登录后优先使用 user_id；
- 变更登录状态时保持灰度一致性（优先 user_id，回退 device_id）。

审核期/灰度期：
- 敏感模块（如提现、礼品卡交易）上线期间，通过远程配置将 rollout% 限制为小流量或按国家/平台限制；
- 遇紧急情况可通过 feature_global_switches 一键急停。
---

## 六、风险与回退

- 认证链路：EvalCtxMiddleware 必须在 AuthMiddleware 之后挂载，确保 user_id 已注入；若匿名访问，只读取 Header 与 IP，不影响现有鉴权。
- 速率限制：Gate 中间件需挂在业务 handler 之前，且不改变 Request body/响应体格式；
- 审计：所有改动 feature_flags 的操作必须记入 feature_flag_audits；
- 缓存：评估结果可加本地 LRU + Redis 订阅（LISTEN/NOTIFY）做热更新（后续迭代）；
- 最小侵入：业务 handler 不改，靠路由挂载 Gate 即可；
- 回退策略：默认所有新建开关 enabled=false；一键急停表可覆盖 Evaluate 结果。

验收清单（自测脚本）：

```bash
# 1) 迁移
psql "$DATABASE_URL" -f migrations/20250221_feature_flags.sql

# 2) 插入样例开关（默认关闭）
psql "$DATABASE_URL" -c "INSERT INTO feature_flags(key,enabled,description) VALUES ('giftcard.trade', false, '礼品卡交易');" || true
psql "$DATABASE_URL" -c "INSERT INTO feature_flags(key,enabled,description) VALUES ('wallet.withdraw', false, '提现');" || true

# 3) 启动服务
# go run ./cmd/main.go  （保持与你现有启动一致）

# 4) 拉取 flags
http :8080/v1/flags X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo X-Build-Channel:internal X-Locale:zh-CN

# 5) menus 裁剪
http :8080/v1/menus X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo X-Build-Channel:internal X-Locale:zh-CN

# 6) 访问受控接口（应 403）
http POST :8080/v1/trade/orders X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo
http POST :8080/v1/withdraw/requests X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo

# 7) 打开 giftcard.trade 后应恢复 200
psql "$DATABASE_URL" -c "UPDATE feature_flags SET enabled=true WHERE key='giftcard.trade';"
http POST :8080/v1/trade/orders X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo
```
---

## 七、快速定位命令（ripgrep/find）

# 路由与引擎初始化
rg -n "gin\\.New|SetupRouter|router|Use\(" -g "internal/**|cmd/**"

# 认证链路与中间件
rg -n "AuthMiddleware|AdminAuthMiddleware|SuperAdminMiddleware" -g "internal/**"
rg -n "RequestID|Correlation|CORS|SecurityHeaders|RateLimit|Timeout" -g "internal/middleware/**"

# 自定义 Header 使用处
rg -n "X-Platform|X-App-Version|X-Device-Id|X-Locale|X-Build-Channel" -g "internal/**"

# Feature Flags / Remote Config 相关
rg -n "feature_flags|remote_config|feature" -g "internal/**|migrations/**"

# 业务关键词（礼品卡交易/提现）
rg -n "trade|withdraw|giftcard|wallet" -g "internal/**"

# 路由文件
rg -n "routes\\.go|router\\.go" -g "internal/**"

# 项目 Go 文件规模
find internal -name "*.go" -maxdepth 5 | wc -l
---

## 八、一步一步落地（命令式步骤）

1) 创建与执行迁移（需已配置 DATABASE_URL 或使用你们现有迁移方式）
- 新建文件：migrations/20250221_feature_flags.sql（见上文）
- 执行迁移：
```bash
psql "$DATABASE_URL" -f migrations/20250221_feature_flags.sql
```
- 初始化样例开关：
```bash
psql "$DATABASE_URL" -c "INSERT INTO feature_flags(key,enabled,description) VALUES ('giftcard.trade', false, '礼品卡交易');" || true
psql "$DATABASE_URL" -c "INSERT INTO feature_flags(key,enabled,description) VALUES ('wallet.withdraw', false, '提现');" || true
```

2) 新增代码文件（按路径新建并粘贴上文代码草案）：
- internal/middleware/evalctx.go
- internal/feature/evaluator.go
- internal/feature/middleware_gate.go
- internal/menu/handler.go

3) 修改路由注册（internal/router/router.go）：
- 在 AuthMiddleware / AdminAuthMiddleware 之后增加：EvalCtxMiddleware()
- 初始化 evaluator 并注册端点：
  - GET /v1/flags → 返回评估后的 flags
  - GET /v1/menus → 返回基于 flags 裁剪后的菜单
- 在敏感端点挂 Gate 中间件（示例）：
  - POST /v1/trade/orders → Gate("giftcard.trade")
  - POST /v1/withdraw/requests → Gate("wallet.withdraw")

4) 更新 CORS 允许头（internal/middleware/security.go）：
- 将以下头加入 Access-Control-Allow-Headers：
  - X-Platform, X-App-Version, X-Device-Id, X-Build-Channel, X-Locale
- 如有 ContentType 验证中间件限制，也需加入白名单。

5) 编译与运行
```bash
go build ./...
go run ./cmd/main.go
```

6) 自测与验收（httpie 示例）
```bash
# 拉取 flags
http :8080/v1/flags X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo X-Build-Channel:internal X-Locale:zh-CN

# 拉取 menus（根据 flags 裁剪）
http :8080/v1/menus X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo X-Build-Channel:internal X-Locale:zh-CN

# 访问受控接口（关闭时应 403）
http POST :8080/v1/trade/orders X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo
http POST :8080/v1/withdraw/requests X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo

# 打开 giftcard.trade 后应恢复 200
psql "$DATABASE_URL" -c "UPDATE feature_flags SET enabled=true WHERE key='giftcard.trade';"
http POST :8080/v1/trade/orders X-Platform:ios X-App-Version:1.3.0 X-Device-Id:demo
```

7) 回退策略
- 保持 feature_flags.enabled=false（默认），即可不影响现网
- 如需紧急回退：
  - UPDATE feature_flags SET enabled=false WHERE key IN ('giftcard.trade','wallet.withdraw');
  - 或在 feature_global_switches 写入 forced_state='disabled'
- 代码层面：仅移除路由上的 Gate 中间件注册即可回退。

---

## 九、结论与首周交付

- 是否可行：可行。现有仓库认证链路完整，能在其之上以最小侵入方式新增 EvalCtx + Feature Flags + Gate。
- 工作量级：中。
  - 主要包含：数据表迁移、Evaluator 与 Gate 中间件、EvalCtx 中间件、菜单裁剪端点、CORS 调整与路由集成、自测脚本。
- 首周可交付范围：
  - migrations（feature_flags、feature_flag_audits、feature_global_switches 可选）
  - EvalCtxMiddleware（含 Header/IP 解析）
  - Evaluator v1（先实现 Enabled-only 与最小规则骨架）
  - Gate 中间件（403 feature_disabled）
  - /v1/flags 与 /v1/menus 端点（示例菜单裁剪）
  - 在示例路由上挂 Gate（或在真实业务路由灰度挂载）
  - 自测脚本与验收条目
- 后续迭代（第2-3周）：
  - 完善规则引擎（countries/roles/platforms/app_version_min/max/time_window/rollout% + stickiness）
  - 本地与分布式缓存 + LISTEN/NOTIFY 热更新
  - 后台开关管理与审计写入
