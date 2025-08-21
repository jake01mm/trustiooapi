package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"trusioo_api/internal/images/dto"
	"trusioo_api/internal/images/entities"
)

type ImageCache struct {
	client     *redis.Client
	expiration time.Duration
}

func NewImageCache(client *redis.Client, expiration time.Duration) *ImageCache {
	return &ImageCache{
		client:     client,
		expiration: expiration,
	}
}

// 缓存键生成
func (c *ImageCache) getUserImageListKey(userID int, folder string, isPublic *bool, page, pageSize int) string {
	publicStr := "all"
	if isPublic != nil {
		publicStr = strconv.FormatBool(*isPublic)
	}
	return fmt.Sprintf("user:%d:images:folder:%s:public:%s:page:%d:size:%d", userID, folder, publicStr, page, pageSize)
}

func (c *ImageCache) getImageKey(imageID int) string {
	return fmt.Sprintf("image:%d", imageID)
}

func (c *ImageCache) getImageByKeyKey(key string) string {
	return fmt.Sprintf("image:key:%s", key)
}

func (c *ImageCache) getAdminImageListKey(folder string, isPublic *bool, page, pageSize int) string {
	publicStr := "all"
	if isPublic != nil {
		publicStr = strconv.FormatBool(*isPublic)
	}
	return fmt.Sprintf("admin:images:folder:%s:public:%s:page:%d:size:%d", folder, publicStr, page, pageSize)
}

// 用户图片列表缓存
func (c *ImageCache) GetUserImageList(ctx context.Context, userID int, folder string, isPublic *bool, page, pageSize int) (*dto.ListImagesResponse, error) {
	key := c.getUserImageListKey(userID, folder, isPublic, page, pageSize)
	
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	
	var response dto.ListImagesResponse
	err = json.Unmarshal([]byte(val), &response)
	return &response, err
}

func (c *ImageCache) SetUserImageList(ctx context.Context, userID int, folder string, isPublic *bool, page, pageSize int, response *dto.ListImagesResponse) error {
	key := c.getUserImageListKey(userID, folder, isPublic, page, pageSize)
	
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	
	return c.client.Set(ctx, key, data, c.expiration).Err()
}

// 单张图片缓存
func (c *ImageCache) GetImage(ctx context.Context, imageID int) (*entities.Image, error) {
	key := c.getImageKey(imageID)
	
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	
	var image entities.Image
	err = json.Unmarshal([]byte(val), &image)
	return &image, err
}

func (c *ImageCache) SetImage(ctx context.Context, image *entities.Image) error {
	key := c.getImageKey(image.ID)
	
	data, err := json.Marshal(image)
	if err != nil {
		return err
	}
	
	return c.client.Set(ctx, key, data, c.expiration).Err()
}

// 通过key获取图片缓存
func (c *ImageCache) GetImageByKey(ctx context.Context, imageKey string) (*entities.Image, error) {
	key := c.getImageByKeyKey(imageKey)
	
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	
	var image entities.Image
	err = json.Unmarshal([]byte(val), &image)
	return &image, err
}

func (c *ImageCache) SetImageByKey(ctx context.Context, imageKey string, image *entities.Image) error {
	key := c.getImageByKeyKey(imageKey)
	
	data, err := json.Marshal(image)
	if err != nil {
		return err
	}
	
	return c.client.Set(ctx, key, data, c.expiration).Err()
}

// 管理员图片列表缓存
func (c *ImageCache) GetAdminImageList(ctx context.Context, folder string, isPublic *bool, page, pageSize int) (*dto.ListImagesResponse, error) {
	key := c.getAdminImageListKey(folder, isPublic, page, pageSize)
	
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	
	var response dto.ListImagesResponse
	err = json.Unmarshal([]byte(val), &response)
	return &response, err
}

func (c *ImageCache) SetAdminImageList(ctx context.Context, folder string, isPublic *bool, page, pageSize int, response *dto.ListImagesResponse) error {
	key := c.getAdminImageListKey(folder, isPublic, page, pageSize)
	
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	
	return c.client.Set(ctx, key, data, c.expiration).Err()
}

// 缓存失效方法
func (c *ImageCache) InvalidateImage(ctx context.Context, imageID int) error {
	key := c.getImageKey(imageID)
	return c.client.Del(ctx, key).Err()
}

func (c *ImageCache) InvalidateImageByKey(ctx context.Context, imageKey string) error {
	key := c.getImageByKeyKey(imageKey)
	return c.client.Del(ctx, key).Err()
}

func (c *ImageCache) InvalidateUserImageLists(ctx context.Context, userID int) error {
	pattern := fmt.Sprintf("user:%d:images:*", userID)
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (c *ImageCache) InvalidateAdminImageLists(ctx context.Context) error {
	pattern := "admin:images:*"
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// 批量失效缓存（当图片被删除时）
func (c *ImageCache) InvalidateAll(ctx context.Context, userID *int) error {
	if userID != nil {
		// 失效用户相关缓存
		if err := c.InvalidateUserImageLists(ctx, *userID); err != nil {
			return err
		}
	}
	
	// 失效管理员缓存（因为管理员可以看到所有图片）
	return c.InvalidateAdminImageLists(ctx)
}

// 预热缓存 - 预加载热点图片
func (c *ImageCache) WarmupPopularImages(ctx context.Context, images []*entities.Image) error {
	for _, image := range images {
		if err := c.SetImage(ctx, image); err != nil {
			// 记录错误但继续处理其他图片
			continue
		}
		
		// 如果是公开图片，也缓存key访问
		if image.IsPublic {
			if err := c.SetImageByKey(ctx, image.Key, image); err != nil {
				continue
			}
		}
	}
	return nil
}

// 获取缓存统计信息
func (c *ImageCache) GetStats(ctx context.Context) map[string]interface{} {
	info, err := c.client.Info(ctx, "stats").Result()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	
	return map[string]interface{}{
		"redis_info": info,
		"expiration": c.expiration.String(),
	}
}