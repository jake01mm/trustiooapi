package middleware

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/imageprocessor"
)

type ImageProcessingConfig struct {
	MaxWidth         uint
	MaxHeight        uint
	Quality          int
	AutoOptimize     bool
	CreateThumbnails bool
	ThumbnailSizes   []ThumbnailSize
}

type ThumbnailSize struct {
	Name   string
	Width  uint
	Height uint
}

func ImageProcessingMiddleware(config ImageProcessingConfig) gin.HandlerFunc {
	processor := imageprocessor.NewProcessor()

	return func(c *gin.Context) {
		if c.Request.Method == "POST" && c.ContentType() == "multipart/form-data" {
			err := c.Request.ParseMultipartForm(32 << 20) // 32 MB
			if err != nil {
				c.JSON(http.StatusBadRequest, common.ErrorResponse{
					Error:   "FORM_PARSE_ERROR",
					Message: "Failed to parse multipart form",
				})
				c.Abort()
				return
			}

			form := c.Request.MultipartForm
			if form == nil || form.File == nil {
				c.Next()
				return
			}

			// Process each file field
			for fieldName, fileHeaders := range form.File {
				if fieldName != "file" && fieldName != "image" {
					continue // Only process specific file fields
				}

				for i, fileHeader := range fileHeaders {
					if !imageprocessor.IsImageFile(fileHeader.Filename) {
						continue // Skip non-image files
					}

					// Get image dimensions
					width, height, err := imageprocessor.GetImageDimensions(fileHeader)
					if err != nil {
						c.JSON(http.StatusBadRequest, common.ErrorResponse{
							Error:   "INVALID_IMAGE",
							Message: "Failed to read image dimensions",
						})
						c.Abort()
						return
					}

					// Check if processing is needed
					needsProcessing := config.AutoOptimize ||
						(config.MaxWidth > 0 && uint(width) > config.MaxWidth) ||
						(config.MaxHeight > 0 && uint(height) > config.MaxHeight)

					if needsProcessing {
						// Process the image
						options := &imageprocessor.ProcessorOptions{
							MaxWidth:  config.MaxWidth,
							MaxHeight: config.MaxHeight,
							Quality:   config.Quality,
							Compress:  config.AutoOptimize,
						}

						processedBuffer, contentType, err := processor.ProcessImage(fileHeader, options)
						if err != nil {
							c.JSON(http.StatusInternalServerError, common.ErrorResponse{
								Error:   "IMAGE_PROCESSING_ERROR",
								Message: "Failed to process image",
							})
							c.Abort()
							return
						}

						// Replace the original file with processed version
						newFileHeader := &multipart.FileHeader{
							Filename: fileHeader.Filename,
							Header:   make(map[string][]string),
							Size:     int64(processedBuffer.Len()),
						}
						newFileHeader.Header.Set("Content-Type", contentType)

						// Create a new file reader from processed buffer
						processedFile := &processedFileReader{
							buffer: processedBuffer,
							pos:    0,
						}

						// Store the processed file in context for later use
						c.Set("processed_file_"+fieldName, processedFile)
						c.Set("processed_file_header_"+fieldName, newFileHeader)

						// Replace in form
						fileHeaders[i] = newFileHeader
					}

					// Create thumbnails if requested
					if config.CreateThumbnails && len(config.ThumbnailSizes) > 0 {
						thumbnails := make(map[string]*bytes.Buffer)

						for _, size := range config.ThumbnailSizes {
							thumbBuffer, _, err := processor.CreateThumbnail(fileHeader, size.Width, size.Height)
							if err == nil {
								thumbnails[size.Name] = thumbBuffer
							}
						}

						if len(thumbnails) > 0 {
							c.Set("thumbnails_"+fieldName, thumbnails)
						}
					}
				}
			}
		}

		c.Next()
	}
}

type processedFileReader struct {
	buffer *bytes.Buffer
	pos    int64
}

func (r *processedFileReader) Read(p []byte) (int, error) {
	if r.pos >= int64(r.buffer.Len()) {
		return 0, io.EOF
	}

	n := copy(p, r.buffer.Bytes()[r.pos:])
	r.pos += int64(n)
	return n, nil
}

func (r *processedFileReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.pos = offset
	case io.SeekCurrent:
		r.pos += offset
	case io.SeekEnd:
		r.pos = int64(r.buffer.Len()) + offset
	}

	if r.pos < 0 {
		r.pos = 0
	}
	if r.pos > int64(r.buffer.Len()) {
		r.pos = int64(r.buffer.Len())
	}

	return r.pos, nil
}

func (r *processedFileReader) Close() error {
	return nil
}

// Helper function to get processed file from context
func GetProcessedFile(c *gin.Context, fieldName string) (multipart.File, *multipart.FileHeader, bool) {
	if file, exists := c.Get("processed_file_" + fieldName); exists {
		if header, headerExists := c.Get("processed_file_header_" + fieldName); headerExists {
			if f, ok := file.(multipart.File); ok {
				if h, ok := header.(*multipart.FileHeader); ok {
					return f, h, true
				}
			}
		}
	}
	return nil, nil, false
}

// Helper function to get thumbnails from context
func GetThumbnails(c *gin.Context, fieldName string) (map[string]*bytes.Buffer, bool) {
	if thumbnails, exists := c.Get("thumbnails_" + fieldName); exists {
		if thumbs, ok := thumbnails.(map[string]*bytes.Buffer); ok {
			return thumbs, true
		}
	}
	return nil, false
}