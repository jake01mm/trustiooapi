package images

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"trusioo_api/internal/images/dto"
	"trusioo_api/internal/images/entities"
	"trusioo_api/pkg/r2storage"
)

type Service interface {
	// 用户接口 - 只能操作自己的图片
	UploadImage(ctx context.Context, userID *int, file *multipart.FileHeader, req dto.UploadImageRequest) (*dto.UploadImageResponse, error)
	GetUserImage(ctx context.Context, userID int, imageID int) (*dto.GetImageResponse, error)
	GetPublicImageByKey(ctx context.Context, key string) (*dto.GetImageResponse, error)
	ListImages(ctx context.Context, userID *int, req dto.ListImagesRequest) (*dto.ListImagesResponse, error)
	DeleteUserImage(ctx context.Context, userID int, imageID int) error
	RefreshUserImageURL(ctx context.Context, userID int, imageID int) (*dto.GetImageResponse, error)
	
	// 管理员接口 - 可以操作所有图片
	AdminListAllImages(ctx context.Context, req dto.ListImagesRequest) (*dto.ListImagesResponse, error)
	AdminGetAnyImage(ctx context.Context, imageID int) (*dto.GetImageResponse, error)
	AdminDeleteAnyImage(ctx context.Context, imageID int) error
	AdminBatchDeleteImages(ctx context.Context, imageIDs []int) (int, error)
}

type service struct {
	repo      Repository
	r2Client  *r2storage.Client
}

func NewService(repo Repository, r2Client *r2storage.Client) Service {
	return &service{
		repo:     repo,
		r2Client: r2Client,
	}
}

// =================== 用户接口 ===================

func (s *service) UploadImage(ctx context.Context, userID *int, file *multipart.FileHeader, req dto.UploadImageRequest) (*dto.UploadImageResponse, error) {
	fileName := req.FileName
	if fileName == "" {
		ext := filepath.Ext(file.Filename)
		fileName = fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	}

	uploadOptions := r2storage.UploadOptions{
		IsPublic: req.IsPublic,
		Folder:   req.Folder,
		FileName: fileName,
	}

	result, err := s.r2Client.UploadFile(ctx, file, uploadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	image := &entities.Image{
		UserID:       userID,
		FileName:     fileName,
		OriginalName: file.Filename,
		Key:          result.Key,
		Bucket:       result.Bucket,
		URL:          result.URL,
		ContentType:  file.Header.Get("Content-Type"),
		Size:         result.Size,
		IsPublic:     req.IsPublic,
	}

	if result.PublicURL != "" {
		image.PublicURL = &result.PublicURL
	}

	if req.Folder != "" {
		image.Folder = &req.Folder
	}

	err = s.repo.Create(ctx, image)
	if err != nil {
		s.r2Client.DeleteFile(ctx, result.Bucket, result.Key)
		return nil, fmt.Errorf("failed to save image metadata: %w", err)
	}

	return &dto.UploadImageResponse{
		ID:           image.ID,
		FileName:     image.FileName,
		OriginalName: image.OriginalName,
		Key:          image.Key,
		URL:          image.URL,
		PublicURL:    image.PublicURL,
		ContentType:  image.ContentType,
		Size:         image.Size,
		IsPublic:     image.IsPublic,
		Folder:       image.Folder,
	}, nil
}

func (s *service) GetUserImage(ctx context.Context, userID int, imageID int) (*dto.GetImageResponse, error) {
	image, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// 验证所有权
	if image.UserID == nil || *image.UserID != userID {
		return nil, fmt.Errorf("access denied: image belongs to another user")
	}

	if !image.IsPublic && image.PublicURL == nil {
		newURL, err := s.r2Client.GeneratePresignedURL(ctx, image.Bucket, image.Key, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
		}
		image.URL = newURL
		s.repo.Update(ctx, image)
	}

	return &dto.GetImageResponse{
		ID:           image.ID,
		FileName:     image.FileName,
		OriginalName: image.OriginalName,
		Key:          image.Key,
		URL:          image.URL,
		PublicURL:    image.PublicURL,
		ContentType:  image.ContentType,
		Size:         image.Size,
		IsPublic:     image.IsPublic,
		Folder:       image.Folder,
		CreatedAt:    image.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) GetPublicImageByKey(ctx context.Context, key string) (*dto.GetImageResponse, error) {
	image, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// 只允许访问公开图片
	if !image.IsPublic {
		return nil, fmt.Errorf("access denied: this is a private image")
	}

	return &dto.GetImageResponse{
		ID:           image.ID,
		FileName:     image.FileName,
		OriginalName: image.OriginalName,
		Key:          image.Key,
		URL:          image.URL,
		PublicURL:    image.PublicURL,
		ContentType:  image.ContentType,
		Size:         image.Size,
		IsPublic:     image.IsPublic,
		Folder:       image.Folder,
		CreatedAt:    image.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) ListImages(ctx context.Context, userID *int, req dto.ListImagesRequest) (*dto.ListImagesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	// 只查询用户自己的图片
	images, total, err := s.repo.List(ctx, userID, req.Folder, req.IsPublic, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	imageResponses := make([]dto.GetImageResponse, len(images))
	for i, image := range images {
		if !image.IsPublic && image.PublicURL == nil {
			newURL, err := s.r2Client.GeneratePresignedURL(ctx, image.Bucket, image.Key, 24*time.Hour)
			if err == nil {
				image.URL = newURL
				s.repo.Update(ctx, image)
			}
		}

		imageResponses[i] = dto.GetImageResponse{
			ID:           image.ID,
			FileName:     image.FileName,
			OriginalName: image.OriginalName,
			Key:          image.Key,
			URL:          image.URL,
			PublicURL:    image.PublicURL,
			ContentType:  image.ContentType,
			Size:         image.Size,
			IsPublic:     image.IsPublic,
			Folder:       image.Folder,
			CreatedAt:    image.CreatedAt.Format(time.RFC3339),
		}
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &dto.ListImagesResponse{
		Images:     imageResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *service) DeleteUserImage(ctx context.Context, userID int, imageID int) error {
	image, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	// 验证所有权
	if image.UserID == nil || *image.UserID != userID {
		return fmt.Errorf("access denied: image belongs to another user")
	}

	err = s.r2Client.DeleteFile(ctx, image.Bucket, image.Key)
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	err = s.repo.Delete(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image from database: %w", err)
	}

	return nil
}

func (s *service) RefreshUserImageURL(ctx context.Context, userID int, imageID int) (*dto.GetImageResponse, error) {
	image, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// 验证所有权
	if image.UserID == nil || *image.UserID != userID {
		return nil, fmt.Errorf("access denied: image belongs to another user")
	}

	if !image.IsPublic {
		newURL, err := s.r2Client.GeneratePresignedURL(ctx, image.Bucket, image.Key, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
		}
		image.URL = newURL
		err = s.repo.Update(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("failed to update image URL: %w", err)
		}
	}

	return &dto.GetImageResponse{
		ID:           image.ID,
		FileName:     image.FileName,
		OriginalName: image.OriginalName,
		Key:          image.Key,
		URL:          image.URL,
		PublicURL:    image.PublicURL,
		ContentType:  image.ContentType,
		Size:         image.Size,
		IsPublic:     image.IsPublic,
		Folder:       image.Folder,
		CreatedAt:    image.CreatedAt.Format(time.RFC3339),
	}, nil
}

// =================== 管理员接口 ===================

func (s *service) AdminListAllImages(ctx context.Context, req dto.ListImagesRequest) (*dto.ListImagesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	// 管理员可以查看所有用户的图片 - 不传userID
	images, total, err := s.repo.List(ctx, nil, req.Folder, req.IsPublic, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	imageResponses := make([]dto.GetImageResponse, len(images))
	for i, image := range images {
		if !image.IsPublic && image.PublicURL == nil {
			newURL, err := s.r2Client.GeneratePresignedURL(ctx, image.Bucket, image.Key, 24*time.Hour)
			if err == nil {
				image.URL = newURL
				s.repo.Update(ctx, image)
			}
		}

		imageResponses[i] = dto.GetImageResponse{
			ID:           image.ID,
			FileName:     image.FileName,
			OriginalName: image.OriginalName,
			Key:          image.Key,
			URL:          image.URL,
			PublicURL:    image.PublicURL,
			ContentType:  image.ContentType,
			Size:         image.Size,
			IsPublic:     image.IsPublic,
			Folder:       image.Folder,
			CreatedAt:    image.CreatedAt.Format(time.RFC3339),
		}
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &dto.ListImagesResponse{
		Images:     imageResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *service) AdminGetAnyImage(ctx context.Context, imageID int) (*dto.GetImageResponse, error) {
	image, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	if !image.IsPublic && image.PublicURL == nil {
		newURL, err := s.r2Client.GeneratePresignedURL(ctx, image.Bucket, image.Key, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
		}
		image.URL = newURL
		s.repo.Update(ctx, image)
	}

	return &dto.GetImageResponse{
		ID:           image.ID,
		FileName:     image.FileName,
		OriginalName: image.OriginalName,
		Key:          image.Key,
		URL:          image.URL,
		PublicURL:    image.PublicURL,
		ContentType:  image.ContentType,
		Size:         image.Size,
		IsPublic:     image.IsPublic,
		Folder:       image.Folder,
		CreatedAt:    image.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) AdminDeleteAnyImage(ctx context.Context, imageID int) error {
	image, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	err = s.r2Client.DeleteFile(ctx, image.Bucket, image.Key)
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	err = s.repo.Delete(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image from database: %w", err)
	}

	return nil
}

func (s *service) AdminBatchDeleteImages(ctx context.Context, imageIDs []int) (int, error) {
	deletedCount := 0
	
	for _, imageID := range imageIDs {
		err := s.AdminDeleteAnyImage(ctx, imageID)
		if err == nil {
			deletedCount++
		}
		// 继续删除其他图片，即使某个删除失败
	}

	return deletedCount, nil
}