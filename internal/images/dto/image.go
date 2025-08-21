package dto

type UploadImageRequest struct {
	IsPublic bool   `form:"is_public"`
	Folder   string `form:"folder"`
	FileName string `form:"file_name"`
}

type UploadImageResponse struct {
	ID          int     `json:"id"`
	FileName    string  `json:"file_name"`
	OriginalName string `json:"original_name"`
	Key         string  `json:"key"`
	URL         string  `json:"url"`
	PublicURL   *string `json:"public_url,omitempty"`
	ContentType string  `json:"content_type"`
	Size        int64   `json:"size"`
	IsPublic    bool    `json:"is_public"`
	Folder      *string `json:"folder,omitempty"`
}

type GetImageResponse struct {
	ID          int     `json:"id"`
	FileName    string  `json:"file_name"`
	OriginalName string `json:"original_name"`
	Key         string  `json:"key"`
	URL         string  `json:"url"`
	PublicURL   *string `json:"public_url,omitempty"`
	ContentType string  `json:"content_type"`
	Size        int64   `json:"size"`
	IsPublic    bool    `json:"is_public"`
	Folder      *string `json:"folder,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type ListImagesRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Folder   string `form:"folder"`
	IsPublic *bool  `form:"is_public"`
}

type ListImagesResponse struct {
	Images     []GetImageResponse `json:"images"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}