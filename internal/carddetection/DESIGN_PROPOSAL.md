# CardDetection 模块重新设计方案

## 概述

本方案将重新设计 `/Users/laitsim/trusioo_api/internal/carddetection` 模块，创建两个独立的数据表来管理产品信息和地区配置，并实现完整的卡片格式验证功能。

## 数据表设计

### 1. 产品表 (cd_products)

```sql
CREATE TABLE cd_products (
    id BIGSERIAL PRIMARY KEY,
    product_mark VARCHAR(20) NOT NULL UNIQUE COMMENT '产品标识',
    product_name VARCHAR(50) NOT NULL COMMENT '产品名称',
    requires_region BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否需要地区',
    requires_pin BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否需要PIN码',
    card_format VARCHAR(100) NOT NULL COMMENT '卡号格式描述',
    card_length_min INTEGER NOT NULL COMMENT '卡号最小长度',
    card_length_max INTEGER NOT NULL COMMENT '卡号最大长度',
    pin_length INTEGER NULL COMMENT 'PIN码长度',
    validation_pattern TEXT NULL COMMENT '验证正则表达式',
    supports_auto_type BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否支持自动识别类型',
    status SMALLINT NOT NULL DEFAULT 1 COMMENT '状态(1=启用 0=禁用)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cd_products_status ON cd_products(status);
CREATE INDEX idx_cd_products_mark ON cd_products(product_mark);
```

### 2. 地区表 (cd_regions)

```sql
CREATE TABLE cd_regions (
    id BIGSERIAL PRIMARY KEY,
    product_mark VARCHAR(20) NOT NULL COMMENT '关联产品标识',
    region_id INTEGER NOT NULL COMMENT '地区ID',
    region_name VARCHAR(50) NOT NULL COMMENT '地区名称',
    region_name_en VARCHAR(50) NULL COMMENT '英文名称',
    status SMALLINT NOT NULL DEFAULT 1 COMMENT '状态(1=启用 0=禁用)',
    sort_order INTEGER NOT NULL DEFAULT 0 COMMENT '排序',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_cd_regions_product FOREIGN KEY (product_mark) REFERENCES cd_products(product_mark) ON DELETE CASCADE,
    CONSTRAINT uk_cd_regions_product_region UNIQUE (product_mark, region_id)
);

CREATE INDEX idx_cd_regions_product ON cd_regions(product_mark);
CREATE INDEX idx_cd_regions_status ON cd_regions(status);
CREATE INDEX idx_cd_regions_sort ON cd_regions(sort_order);
```

## 基础数据设计

### 产品数据 (7种产品类型)

| product_mark | product_name | requires_region | requires_pin | card_format | card_length_min | card_length_max | pin_length | validation_pattern | supports_auto_type |
|--------------|--------------|-----------------|--------------|-------------|-----------------|-----------------|------------|-------------------|--------------------|
| sephora | 丝芙兰 | false | true | 16位卡号-8位PIN | 25 | 25 | 8 | ^[0-9]{16}-[0-9]{8}$ | false |
| Razer | 雷蛇 | true | false | 标准卡号 | 10 | 20 | null | ^[A-Z0-9]+$ | false |
| iTunes | 苹果 | true | false | 16位字符 | 16 | 16 | null | ^[A-Z0-9]{16}$ | true |
| amazon | 亚马逊 | true | false | 14/15位数字 | 14 | 15 | null | ^[0-9]{14,15}$ | false |
| xBox | XBOX | true | false | 25位字符 | 25 | 25 | null | ^[A-Z0-9]{25}$ | false |
| nike | NIKE | false | true | 19位卡号-6位PIN | 26 | 26 | 6 | ^[0-9]{19}-[0-9]{6}$ | false |
| nd | ND | false | true | 16位卡号-8位PIN | 25 | 25 | 8 | ^[0-9]{16}-[0-9]{8}$ | false |

### 地区数据 (总计51条记录)

#### iTunes地区 (11条)
```
{1, "英国"}, {2, "美国"}, {3, "德国"}, {4, "澳大利亚"},
{5, "加拿大"}, {6, "日本"}, {8, "西班牙"}, {9, "意大利"},
{10, "法国"}, {11, "爱尔兰"}, {12, "墨西哥"}
```

#### Amazon地区 (2条)
```
{2, "美亚/加亚"}, {1, "欧盟区"}
```

#### Razer地区 (22条)
```
{12, "美国"}, {6, "澳大利亚"}, {13, "巴西"}, {26, "柬埔寨"},
{20, "加拿大"}, {25, "智利"}, {22, "哥伦比亚"}, {17, "香港特别行政区"},
{4, "印度"}, {7, "印度尼西亚"}, {27, "日本"}, {1, "马来西亚"},
{19, "缅甸"}, {15, "新西兰"}, {29, "巴基斯坦"}, {8, "菲律宾"},
{5, "新加坡"}, {18, "土耳其"}, {33, "越南"}, {2, "其他"},
{28, "其他（中文）"}, {21, "墨西哥"}
```

#### Xbox地区 (16条)
```
"美国", "加拿大", "英国", "澳大利亚", "新西兰", "新加坡",
"韩国", "墨西哥", "瑞典", "哥伦比亚", "阿根廷", "尼日利亚",
"香港特别行政区", "挪威", "波兰", "德国"
```

## 卡片格式验证功能

### 验证规则设计

1. **iTunes**: `^[A-Z0-9]{16}$` - 16位大写字母和数字
2. **Amazon**: `^[0-9]{14,15}$` - 14或15位数字
3. **Xbox**: `^[A-Z0-9]{25}$` - 25位大写字母和数字
4. **Nike**: `^[0-9]{19}-[0-9]{6}$` - 19位数字-6位数字
5. **Sephora**: `^[0-9]{16}-[0-9]{8}$` - 16位数字-8位数字
6. **Razer**: `^[A-Z0-9]+$` - 大写字母和数字组合
7. **ND**: `^[0-9]{16}-[0-9]{8}$` - 16位数字-8位数字

### 验证逻辑实现

```go
// 在service.go中实现
func (s *Service) ValidateCardFormat(productMark string, cardNo string) error {
    product, err := s.repo.GetProductByMark(productMark)
    if err != nil {
        return err
    }
    
    if product.ValidationPattern != nil {
        matched, err := regexp.MatchString(*product.ValidationPattern, cardNo)
        if err != nil {
            return fmt.Errorf("validation pattern error: %v", err)
        }
        if !matched {
            return fmt.Errorf("card format invalid for product %s", productMark)
        }
    }
    
    return nil
}
```

## 实体类设计

### entities/product.go

```go
package entities

import "time"

type Product struct {
    ID                int64     `json:"id" db:"id"`
    ProductMark       string    `json:"product_mark" db:"product_mark"`
    ProductName       string    `json:"product_name" db:"product_name"`
    RequiresRegion    bool      `json:"requires_region" db:"requires_region"`
    RequiresPin       bool      `json:"requires_pin" db:"requires_pin"`
    CardFormat        string    `json:"card_format" db:"card_format"`
    CardLengthMin     int       `json:"card_length_min" db:"card_length_min"`
    CardLengthMax     int       `json:"card_length_max" db:"card_length_max"`
    PinLength         *int      `json:"pin_length" db:"pin_length"`
    ValidationPattern *string   `json:"validation_pattern" db:"validation_pattern"`
    SupportsAutoType  bool      `json:"supports_auto_type" db:"supports_auto_type"`
    Status            int       `json:"status" db:"status"`
    CreatedAt         time.Time `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
```

### entities/region.go

```go
package entities

import "time"

type Region struct {
    ID           int64     `json:"id" db:"id"`
    ProductMark  string    `json:"product_mark" db:"product_mark"`
    RegionID     int       `json:"region_id" db:"region_id"`
    RegionName   string    `json:"region_name" db:"region_name"`
    RegionNameEn *string   `json:"region_name_en" db:"region_name_en"`
    Status       int       `json:"status" db:"status"`
    SortOrder    int       `json:"sort_order" db:"sort_order"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
```

## 实施计划

### 第一阶段：数据库迁移

1. **创建迁移文件**
   ```bash
   cd /Users/laitsim/trusioo_api/tools
   make db-create NAME=create_carddetection_tables
   ```

2. **编写迁移SQL**
   - 创建cd_products表
   - 创建cd_regions表
   - 插入7种产品的基础数据
   - 插入51条地区数据

3. **执行迁移**
   ```bash
   make db-migrate
   ```

### 第二阶段：代码重构

1. **创建实体类**
   - `entities/product.go`
   - `entities/region.go`

2. **创建DTO**
   - `dto/product_dto.go`
   - `dto/region_dto.go`

3. **更新Repository**
   - 添加产品查询方法
   - 添加地区查询方法

4. **更新Service**
   - 添加格式验证逻辑
   - 更新业务逻辑

5. **更新Handler**
   - 更新接口实现
   - 添加验证调用

### 第三阶段：测试验证

1. **单元测试**
   - 格式验证测试
   - 数据查询测试

2. **集成测试**
   - API接口测试
   - 数据库操作测试

3. **Postman测试**
   - 更新现有测试集合
   - 验证新功能

## 优势分析

### 1. 数据独立性
- 两个表仅限carddetection模块使用
- 不与其他模块产生依赖关系
- 便于模块独立维护和升级

### 2. 扩展性
- 支持新产品类型的快速添加
- 支持地区配置的灵活管理
- 支持验证规则的动态配置

### 3. 数据完整性
- 外键约束确保数据一致性
- 唯一约束防止重复数据
- 状态字段支持软删除

### 4. 性能优化
- 合理的索引设计
- 支持高效的查询操作
- 减少不必要的数据传输

## 风险评估

### 1. 迁移风险
- **风险**: 数据迁移可能影响现有功能
- **缓解**: 在测试环境充分验证后再部署

### 2. 兼容性风险
- **风险**: 现有API可能需要调整
- **缓解**: 保持API接口不变，仅修改内部实现

### 3. 性能风险
- **风险**: 新增数据库查询可能影响性能
- **缓解**: 合理设计索引，使用缓存机制

## 总结

本设计方案严格遵循用户要求，创建了两个独立的数据表来管理产品和地区信息，实现了完整的卡片格式验证功能。方案具有良好的扩展性、维护性和性能表现，能够满足carddetection模块的长期发展需求。

**请审阅此设计方案，确认无误后我将开始实施开发工作。**