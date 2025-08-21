# TODO: Redis验证码集成

## 🎯 任务概述
在项目开发完成后，将验证码系统从PostgreSQL迁移到Redis缓存，以提升高并发性能。

## 📋 详细任务列表

### Phase 1: 准备工作
- [ ] 备份当前的验证码实现
- [ ] 创建Redis和PostgreSQL双写模式（渐进式迁移）
- [ ] 编写数据迁移脚本

### Phase 2: 核心功能替换

#### 2.1 修改 `internal/auth/verification/service.go`
- [ ] 集成 `pkg/redis/verification.go`
- [ ] 替换 `SendVerificationCode` 实现
```go
// 当前: 直接存储到PostgreSQL
err = s.repo.CreateVerification(verification)

// TODO: 改为Redis + 频率限制
vc := redis.NewVerificationCache()
if blocked, _ := vc.CheckSendFrequency(req.Target, req.Type, 60*time.Second); blocked {
    return nil, errors.New("发送过于频繁，请稍后重试")
}
err = vc.StoreVerificationCode(req.Target, req.Type, code, 10*time.Minute)
vc.SetSendFrequency(req.Target, req.Type, 60*time.Second)
```

#### 2.2 修改 `VerifyCode` 实现
- [ ] 替换验证逻辑为Redis查询
- [ ] 实现失败次数限制
```go
// 当前: 查询PostgreSQL + 标记使用
verification, err := s.repo.GetValidVerification(req.Target, req.Type, req.Code)

// TODO: 改为Redis原子操作
vc := redis.NewVerificationCache()
if blocked, _ := vc.IsBlocked(req.Target, req.Type, 5); blocked {
    return nil, errors.New("验证失败次数过多，请稍后重试")
}

storedCode, err := vc.GetVerificationCode(req.Target, req.Type)
if storedCode != req.Code {
    vc.IncrementAttemptCount(req.Target, req.Type, 1*time.Hour)
    return nil, errors.New("验证码错误")
}

// 验证成功，清理
vc.DeleteVerificationCode(req.Target, req.Type)
vc.ClearAttemptCount(req.Target, req.Type)
```

### Phase 3: 性能优化

#### 3.1 bcrypt优化
- [ ] 调整bcrypt cost从DefaultCost(10)到8
- [ ] 实现密码哈希的异步处理
```go
// internal/auth/user_auth/service.go
const OptimalCost = 8  // 约25ms，平衡安全性和性能

func (s *Service) hashPasswordAsync(password string) <-chan hashResult {
    resultChan := make(chan hashResult, 1)
    go func() {
        hash, err := bcrypt.GenerateFromPassword([]byte(password), OptimalCost)
        resultChan <- hashResult{hash: string(hash), err: err}
    }()
    return resultChan
}
```

#### 3.2 数据库原子性优化
- [ ] 实现注册操作的分布式锁
- [ ] 优化邮箱唯一性检查
```go
func (s *Service) RegisterWithLock(req *dto.RegisterRequest) error {
    vc := redis.NewVerificationCache()
    lockKey := fmt.Sprintf("reg_lock:%s", req.Email)
    
    // 使用Redis实现分布式锁
    locked, err := vc.client.SetNX(context.Background(), lockKey, "1", 5*time.Second).Result()
    if err != nil || !locked {
        return errors.New("注册繁忙，请稍后重试")
    }
    defer vc.client.Del(context.Background(), lockKey)
    
    // 在锁保护下执行注册
    return s.doRegister(req)
}
```

### Phase 4: 监控和测试

#### 4.1 添加监控指标
- [ ] 创建Prometheus指标
- [ ] 监控Redis连接和性能
- [ ] 验证码相关指标统计

#### 4.2 压力测试
- [ ] 编写高并发测试脚本
- [ ] 验证频率限制功能
- [ ] 测试故障恢复能力

#### 4.3 兼容性测试
- [ ] 确保现有API行为不变
- [ ] 验证错误信息一致性
- [ ] 检查Postman集合兼容性

### Phase 5: 部署和回滚

#### 5.1 部署策略
- [ ] 蓝绿部署配置
- [ ] 健康检查更新
- [ ] Redis配置优化

#### 5.2 回滚方案
- [ ] 快速切换回PostgreSQL
- [ ] 数据同步验证
- [ ] 性能回归测试

## 📊 预期性能提升

| 指标 | 当前(PostgreSQL) | 目标(Redis) | 提升比例 |
|------|-----------------|-------------|----------|
| 验证码发送QPS | ~10 | ~500 | 50倍 |
| 验证码验证QPS | ~50 | ~1000 | 20倍 |
| 平均响应时间 | 200ms | 20ms | 10倍 |
| P99响应时间 | 2000ms | 200ms | 10倍 |
| 数据库负载 | 高 | 低 | -80% |

## ⚠️ 风险控制

### 高风险项
- [ ] Redis服务可用性依赖
- [ ] 数据持久化策略
- [ ] 缓存穿透保护

### 缓解措施
- [ ] 实现Redis + PostgreSQL降级机制
- [ ] 配置Redis持久化（RDB + AOF）
- [ ] 添加熔断器模式

## 🔧 开发环境准备

### 必要组件
- [x] Redis服务已启动
- [x] 连接池已配置
- [x] 健康检查已实现
- [x] 基础缓存API已就绪

### 配置文件
- [x] `.env` Redis配置已启用
- [x] `config/config.go` Redis解析正常
- [x] `pkg/redis/` 目录结构完整

## 📅 建议执行时间
**项目开发阶段完成后，生产部署前2-3周开始集成**

## 📝 集成检查清单
执行前确认：
- [ ] 所有核心功能开发完成
- [ ] API测试套件通过率100%
- [ ] 现有验证码功能稳定运行
- [ ] Redis服务生产就绪
- [ ] 监控告警配置完成
- [ ] 回滚方案验证通过

---

**注意：此集成将显著提升系统并发能力，但需要在项目稳定后谨慎执行。**