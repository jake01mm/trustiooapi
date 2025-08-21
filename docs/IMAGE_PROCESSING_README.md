# Trusioo API 图片处理模块使用指南

## 功能概述

本模块提供完整的图片存储和处理功能，支持：

- ✅ 图片上传到 Cloudflare R2 存储桶
- ✅ 公有和私有图片存储
- ✅ CDN 加速访问
- ✅ 图片压缩和格式转换
- ✅ 缩略图生成
- ✅ 预签名 URL 访问私有图片
- ✅ 高并发和高性能支持

## 配置说明

### 1. 环境变量配置

在 `.env` 文件中配置以下 R2 存储相关参数：

```env
# R2 存储配置
R2_ACCESS_KEY_ID=your_r2_access_key_id
R2_SECRET_ACCESS_KEY=your_r2_secret_access_key
R2_ENDPOINT=https://27f7f20b92ac245bf54ced4369c47776.r2.cloudflarestorage.com
R2_REGION=auto
R2_PUBLIC_BUCKET=trusioo-public
R2_PRIVATE_BUCKET=trusioo-private3235
R2_PUBLIC_CDN_URL=https://trusioo-public.trusioo.com
R2_PRIVATE_CDN_URL=https://trusioo-private.trusioo.com
R2_MAX_FILE_SIZE=10485760
R2_ALLOWED_MIME_TYPES=image/jpeg,image/png,image/gif,image/webp
```

### 2. 数据库迁移

运行数据库迁移创建 images 表：

```bash
make migrate-up
# 或者
go run tools/db/migrate/main.go -direction=up
```

## API 端点

### 1. 上传图片

**POST** `/api/v1/images/upload`

**参数：**
- `file` (文件): 要上传的图片文件
- `is_public` (布尔值): 是否为公开图片，默认 false
- `folder` (字符串): 存储文件夹，可选
- `file_name` (字符串): 自定义文件名，可选

**示例：**
```bash
curl -X POST http://localhost:8080/api/v1/images/upload \\
  -F "file=@example.jpg" \\
  -F "is_public=true" \\
  -F "folder=uploads"
```

**响应：**
```json
{
  "message": "Image uploaded successfully",
  "data": {
    "id": 1,
    "file_name": "1645678901234567890.jpg",
    "original_name": "example.jpg",
    "key": "uploads/1645678901234567890.jpg",
    "url": "https://trusioo-public.trusioo.com/uploads/1645678901234567890.jpg",
    "public_url": "https://trusioo-public.trusioo.com/uploads/1645678901234567890.jpg",
    "content_type": "image/jpeg",
    "size": 245760,
    "is_public": true,
    "folder": "uploads"
  }
}
```

### 2. 获取图片列表

**GET** `/api/v1/images/`

**查询参数：**
- `page` (整数): 页码，默认 1
- `page_size` (整数): 每页数量，默认 20，最大 100
- `folder` (字符串): 筛选文件夹
- `is_public` (布尔值): 筛选公开/私有图片

**示例：**
```bash
curl "http://localhost:8080/api/v1/images/?page=1&page_size=10&folder=uploads"
```

### 3. 获取图片详情

**GET** `/api/v1/images/{id}`

**示例：**
```bash
curl http://localhost:8080/api/v1/images/1
```

### 4. 通过 Key 获取公开图片

**GET** `/api/v1/images/public/{key}`

**示例：**
```bash
curl http://localhost:8080/api/v1/images/public/uploads/example.jpg
```

### 5. 刷新私有图片 URL

**PUT** `/api/v1/images/{id}/refresh`

用于重新生成私有图片的访问 URL（24小时有效期）。

### 6. 删除图片

**DELETE** `/api/v1/images/{id}`

同时从 R2 存储桶和数据库中删除图片。

## 图片处理功能

### 1. 自动图片优化

系统会自动对上传的图片进行以下处理：

- 图片尺寸限制：最大宽度 2048px，最大高度 2048px
- JPEG 压缩：质量 85%
- 格式优化：PNG 无透明度时转换为 JPEG
- 文件大小限制：最大 10MB

### 2. 图片处理中间件

可以在路由中使用图片处理中间件：

```go
imageConfig := middleware.ImageProcessingConfig{
    MaxWidth:         1920,
    MaxHeight:        1920,
    Quality:          85,
    AutoOptimize:     true,
    CreateThumbnails: true,
    ThumbnailSizes: []middleware.ThumbnailSize{
        {Name: "small", Width: 150, Height: 150},
        {Name: "medium", Width: 300, Height: 300},
        {Name: "large", Width: 800, Height: 600},
    },
}

r.Use(middleware.ImageProcessingMiddleware(imageConfig))
```

## 存储桶配置

### 1. 公有存储桶 (trusioo-public)

- 配置了公共读取权限
- 文件可通过 CDN 直接访问
- 适用于网站图片、头像等公开内容

### 2. 私有存储桶 (trusioo-private3235)

- 需要预签名 URL 访问
- 提供更高安全性
- 适用于用户私人图片、文档等敏感内容

## CDN 配置

### 1. 自定义域名

- 公有桶：`https://trusioo-public.trusioo.com`
- 私有桶：`https://trusioo-private.trusioo.com`

### 2. 缓存优化

CDN 会自动缓存静态图片，提供全球加速访问。

## 安全特性

### 1. 文件类型验证

只允许以下图片格式：
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)

### 2. 文件大小限制

- 默认最大文件大小：10MB
- 可通过环境变量 `R2_MAX_FILE_SIZE` 配置

### 3. MIME 类型检查

上传时会验证文件的实际 MIME 类型，防止恶意文件上传。

### 4. 预签名 URL

私有图片使用预签名 URL 访问，提供时限性安全访问。

## 使用示例

### 1. 前端上传示例 (JavaScript)

```javascript
async function uploadImage(file, isPublic = true, folder = 'uploads') {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('is_public', isPublic.toString());
    formData.append('folder', folder);

    try {
        const response = await fetch('/api/v1/images/upload', {
            method: 'POST',
            body: formData,
            headers: {
                // 'Authorization': 'Bearer ' + token // 如果需要认证
            }
        });

        const result = await response.json();
        console.log('上传成功:', result);
        return result.data;
    } catch (error) {
        console.error('上传失败:', error);
        throw error;
    }
}
```

### 2. React 组件示例

```jsx
import React, { useState } from 'react';

function ImageUploader() {
    const [uploading, setUploading] = useState(false);
    const [imageUrl, setImageUrl] = useState('');

    const handleFileSelect = async (event) => {
        const file = event.target.files[0];
        if (!file) return;

        setUploading(true);
        try {
            const result = await uploadImage(file, true, 'avatars');
            setImageUrl(result.url);
        } catch (error) {
            alert('上传失败: ' + error.message);
        } finally {
            setUploading(false);
        }
    };

    return (
        <div>
            <input 
                type="file" 
                accept="image/*" 
                onChange={handleFileSelect}
                disabled={uploading}
            />
            {uploading && <p>上传中...</p>}
            {imageUrl && (
                <img src={imageUrl} alt="Uploaded" style={{maxWidth: '300px'}} />
            )}
        </div>
    );
}
```

## 测试

### 1. 运行测试脚本

```bash
# 启动服务器
go run cmd/main.go

# 在另一个终端运行测试
node test_image_upload.js
```

### 2. 使用 Postman

导入 `Postman/Image_Management_Module.postman_collection.json` 集合进行测试。

### 3. 手动测试

```bash
# 测试上传
curl -X POST http://localhost:8080/api/v1/images/upload \\
  -F "file=@test.jpg" \\
  -F "is_public=true" \\
  -F "folder=test"

# 测试列表
curl http://localhost:8080/api/v1/images/
```

## 性能优化建议

### 1. CDN 配置

- 启用 Cloudflare 缓存
- 设置适当的缓存策略
- 使用 WebP 格式以获得更好压缩

### 2. 数据库优化

- 定期清理过期的临时文件记录
- 为查询字段添加适当索引

### 3. 存储优化

- 设置生命周期策略自动清理过期文件
- 使用 R2 的分析功能监控使用情况

## 故障排除

### 1. 上传失败

- 检查 R2 凭证配置
- 验证网络连接
- 确认文件格式和大小限制

### 2. 图片无法访问

- 检查 CDN 配置
- 验证存储桶权限
- 确认预签名 URL 未过期

### 3. 性能问题

- 监控 R2 使用情况
- 检查数据库连接池配置
- 优化图片处理参数

## 支持与维护

如有问题请查看：
1. 服务器日志：`logs/app.log`
2. 数据库连接状态
3. R2 存储桶配置
4. 网络连通性

更多详细信息请参考项目文档或联系开发团队。