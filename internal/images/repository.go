package images

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"trusioo_api/internal/images/entities"
)

type Repository interface {
	Create(ctx context.Context, image *entities.Image) error
	GetByID(ctx context.Context, id int) (*entities.Image, error)
	GetByKey(ctx context.Context, key string) (*entities.Image, error)
	List(ctx context.Context, userID *int, folder string, isPublic *bool, offset, limit int) ([]*entities.Image, int64, error)
	Update(ctx context.Context, image *entities.Image) error
	Delete(ctx context.Context, id int) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, image *entities.Image) error {
	query := `
		INSERT INTO images (user_id, file_name, original_name, key, bucket, url, public_url, content_type, size, is_public, folder, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		image.UserID,
		image.FileName,
		image.OriginalName,
		image.Key,
		image.Bucket,
		image.URL,
		image.PublicURL,
		image.ContentType,
		image.Size,
		image.IsPublic,
		image.Folder,
	).Scan(&image.ID, &image.CreatedAt, &image.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int) (*entities.Image, error) {
	query := `
		SELECT id, user_id, file_name, original_name, key, bucket, url, public_url, content_type, size, is_public, folder, created_at, updated_at
		FROM images 
		WHERE id = $1`
	
	image := &entities.Image{}
	err := r.db.GetContext(ctx, image, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get image by ID: %w", err)
	}
	
	return image, nil
}

func (r *repository) GetByKey(ctx context.Context, key string) (*entities.Image, error) {
	query := `
		SELECT id, user_id, file_name, original_name, key, bucket, url, public_url, content_type, size, is_public, folder, created_at, updated_at
		FROM images 
		WHERE key = $1`
	
	image := &entities.Image{}
	err := r.db.GetContext(ctx, image, query, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get image by key: %w", err)
	}
	
	return image, nil
}

func (r *repository) List(ctx context.Context, userID *int, folder string, isPublic *bool, offset, limit int) ([]*entities.Image, int64, error) {
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}
	
	if folder != "" {
		conditions = append(conditions, fmt.Sprintf("folder = $%d", argIndex))
		args = append(args, folder)
		argIndex++
	}
	
	if isPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *isPublic)
		argIndex++
	}
	
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%s", conditions[0])
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}
	
	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM images %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count images: %w", err)
	}
	
	// Get images with pagination
	query := fmt.Sprintf(`
		SELECT id, user_id, file_name, original_name, key, bucket, url, public_url, content_type, size, is_public, folder, created_at, updated_at
		FROM images %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	images := []*entities.Image{}
	err = r.db.SelectContext(ctx, &images, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list images: %w", err)
	}
	
	return images, total, nil
}

func (r *repository) Update(ctx context.Context, image *entities.Image) error {
	query := `
		UPDATE images 
		SET url = $2, public_url = $3, updated_at = NOW()
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, image.ID, image.URL, image.PublicURL)
	if err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}
	
	return nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM images WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("image not found")
	}
	
	return nil
}