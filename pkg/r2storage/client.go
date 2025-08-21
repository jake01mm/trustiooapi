package r2storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type Client struct {
	s3Client      *s3.Client
	publicBucket  string
	privateBucket string
	publicCDNURL  string
	privateCDNURL string
	maxFileSize   int64
	allowedMimeTypes []string
}

type UploadOptions struct {
	IsPublic    bool
	Folder      string
	FileName    string
	ContentType string
}

type UploadResult struct {
	Key       string `json:"key"`
	URL       string `json:"url"`
	PublicURL string `json:"public_url,omitempty"`
	Size      int64  `json:"size"`
	Bucket    string `json:"bucket"`
}

func NewClient(accessKeyID, secretAccessKey, endpoint, region, publicBucket, privateBucket, publicCDNURL, privateCDNURL string, maxFileSize int64, allowedMimeTypes []string) *Client {
	cfg := aws.Config{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}),
	}

	s3Client := s3.NewFromConfig(cfg)

	return &Client{
		s3Client:      s3Client,
		publicBucket:  publicBucket,
		privateBucket: privateBucket,
		publicCDNURL:  publicCDNURL,
		privateCDNURL: privateCDNURL,
		maxFileSize:   maxFileSize,
		allowedMimeTypes: allowedMimeTypes,
	}
}

func (c *Client) ValidateFile(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > c.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", fileHeader.Size, c.maxFileSize)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentType := http.DetectContentType(buffer)
	
	if !slices.Contains(c.allowedMimeTypes, contentType) {
		return fmt.Errorf("file type %s is not allowed", contentType)
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, options UploadOptions) (*UploadResult, error) {
	if err := c.ValidateFile(fileHeader); err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	bucket := c.publicBucket
	cdnURL := c.publicCDNURL
	if !options.IsPublic {
		bucket = c.privateBucket
		cdnURL = c.privateCDNURL
	}

	fileName := options.FileName
	if fileName == "" {
		ext := filepath.Ext(fileHeader.Filename)
		fileName = uuid.New().String() + ext
	}

	key := fileName
	if options.Folder != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(options.Folder, "/"), fileName)
	}

	contentType := options.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(fileContent[:512])
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(fileContent),
		ContentLength: aws.Int64(fileHeader.Size),
		ContentType:   aws.String(contentType),
	}

	if options.IsPublic {
		input.ACL = types.ObjectCannedACLPublicRead
	}

	_, err = c.s3Client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	result := &UploadResult{
		Key:    key,
		Size:   fileHeader.Size,
		Bucket: bucket,
	}

	if options.IsPublic && cdnURL != "" {
		result.PublicURL = fmt.Sprintf("%s/%s", strings.TrimRight(cdnURL, "/"), key)
		result.URL = result.PublicURL
	} else {
		signedURL, err := c.GeneratePresignedURL(ctx, bucket, key, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signed URL: %w", err)
		}
		result.URL = signedURL
	}

	return result, nil
}

func (c *Client) GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3Client)
	
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

func (c *Client) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	return nil
}

func (c *Client) GetFileInfo(ctx context.Context, bucket, key string) (*s3.HeadObjectOutput, error) {
	result, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return result, nil
}