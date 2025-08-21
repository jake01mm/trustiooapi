package imageprocessor

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

type ProcessorOptions struct {
	MaxWidth    uint
	MaxHeight   uint
	Quality     int // For JPEG compression (1-100)
	Format      string // Target format: "jpeg", "png", "webp"
	Compress    bool
}

type Processor struct {
	defaultOptions ProcessorOptions
}

func NewProcessor() *Processor {
	return &Processor{
		defaultOptions: ProcessorOptions{
			MaxWidth:  2048,
			MaxHeight: 2048,
			Quality:   85,
			Format:    "", // Keep original format
			Compress:  true,
		},
	}
}

func (p *Processor) ProcessImage(fileHeader *multipart.FileHeader, options *ProcessorOptions) (*bytes.Buffer, string, error) {
	if options == nil {
		options = &p.defaultOptions
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	// Decode image
	img, originalFormat, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Determine target format
	targetFormat := originalFormat
	if options.Format != "" {
		targetFormat = options.Format
	}

	// Resize if necessary
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	if (options.MaxWidth > 0 && width > options.MaxWidth) || 
	   (options.MaxHeight > 0 && height > options.MaxHeight) {
		img = resize.Thumbnail(options.MaxWidth, options.MaxHeight, img, resize.Lanczos3)
	}

	// Encode to target format
	var output bytes.Buffer
	var contentType string

	switch targetFormat {
	case "jpeg", "jpg":
		quality := options.Quality
		if quality == 0 {
			quality = 85
		}
		err = jpeg.Encode(&output, img, &jpeg.Options{Quality: quality})
		contentType = "image/jpeg"
	case "png":
		err = png.Encode(&output, img)
		contentType = "image/png"
	case "gif":
		err = gif.Encode(&output, img, &gif.Options{})
		contentType = "image/gif"
	default:
		// Keep original format
		switch originalFormat {
		case "jpeg":
			quality := options.Quality
			if quality == 0 {
				quality = 85
			}
			err = jpeg.Encode(&output, img, &jpeg.Options{Quality: quality})
			contentType = "image/jpeg"
		case "png":
			err = png.Encode(&output, img)
			contentType = "image/png"
		case "gif":
			err = gif.Encode(&output, img, &gif.Options{})
			contentType = "image/gif"
		default:
			return nil, "", fmt.Errorf("unsupported image format: %s", originalFormat)
		}
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to encode image: %w", err)
	}

	return &output, contentType, nil
}

func (p *Processor) CreateThumbnail(fileHeader *multipart.FileHeader, width, height uint) (*bytes.Buffer, string, error) {
	options := &ProcessorOptions{
		MaxWidth:  width,
		MaxHeight: height,
		Quality:   80,
		Compress:  true,
	}

	return p.ProcessImage(fileHeader, options)
}

func (p *Processor) CompressImage(fileHeader *multipart.FileHeader, quality int) (*bytes.Buffer, string, error) {
	options := &ProcessorOptions{
		MaxWidth:  p.defaultOptions.MaxWidth,
		MaxHeight: p.defaultOptions.MaxHeight,
		Quality:   quality,
		Compress:  true,
	}

	return p.ProcessImage(fileHeader, options)
}

func (p *Processor) ConvertFormat(fileHeader *multipart.FileHeader, targetFormat string) (*bytes.Buffer, string, error) {
	options := &ProcessorOptions{
		Format:  targetFormat,
		Quality: 85,
	}

	return p.ProcessImage(fileHeader, options)
}

func GetImageDimensions(fileHeader *multipart.FileHeader) (int, int, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}

	return img.Width, img.Height, nil
}

func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return true
	default:
		return false
	}
}

func GetOptimizedFormat(originalFormat string, hasTransparency bool) string {
	switch originalFormat {
	case "png":
		if !hasTransparency {
			return "jpeg" // Convert PNG without transparency to JPEG for better compression
		}
		return "png"
	case "gif":
		return "png" // Convert GIF to PNG for better quality
	case "bmp":
		return "jpeg" // Convert BMP to JPEG
	default:
		return originalFormat
	}
}