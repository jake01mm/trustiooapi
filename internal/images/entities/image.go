package entities

import (
	"time"
)

type Image struct {
	ID          int       `db:"id" json:"id"`
	UserID      *int      `db:"user_id" json:"user_id,omitempty"`
	FileName    string    `db:"file_name" json:"file_name"`
	OriginalName string   `db:"original_name" json:"original_name"`
	Key         string    `db:"key" json:"key"`
	Bucket      string    `db:"bucket" json:"bucket"`
	URL         string    `db:"url" json:"url"`
	PublicURL   *string   `db:"public_url" json:"public_url,omitempty"`
	ContentType string    `db:"content_type" json:"content_type"`
	Size        int64     `db:"size" json:"size"`
	IsPublic    bool      `db:"is_public" json:"is_public"`
	Folder      *string   `db:"folder" json:"folder,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}