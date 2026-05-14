package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"nutrilens/config"

	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// Client wraps the COS SDK client.
type Client struct {
	client    *cos.Client
	bucket    string
	region    string
	baseURL   string
	secretID  string
	secretKey string
}

// NewClient creates a new COS client from config.
func NewClient(cfg config.COSConfig) (*Client, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("parse bucket url: %w", err)
	}

	serviceURL, err := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("parse service url: %w", err)
	}

	client := cos.NewClient(&cos.BaseURL{
		BucketURL:  bucketURL,
		ServiceURL: serviceURL,
	}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	baseURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region)

	return &Client{
		client:    client,
		bucket:    cfg.Bucket,
		region:    cfg.Region,
		baseURL:   baseURL,
		secretID:  cfg.SecretID,
		secretKey: cfg.SecretKey,
	}, nil
}

// UploadResult contains the uploaded object info.
type UploadResult struct {
	ObjectKey string `json:"object_key"`
	ObjectURL string `json:"object_url"`
}

// Upload reads from reader and uploads to COS at {bizType}/{uuid}_{filename}.
func (c *Client) Upload(ctx context.Context, bizType BizType, filename string, reader io.Reader) (*UploadResult, error) {
	if !ValidBizTypes()[bizType] {
		return nil, fmt.Errorf("invalid biz_type: %s", bizType)
	}

	objectKey := fmt.Sprintf("%s/%s_%s", bizType, uuid.New().String(), filename)

	_, err := c.client.Object.Put(ctx, objectKey, reader, nil)
	if err != nil {
		return nil, fmt.Errorf("upload to cos: %w", err)
	}

	return &UploadResult{
		ObjectKey: objectKey,
		ObjectURL: c.baseURL + "/" + objectKey,
	}, nil
}

// GetObject downloads an object from COS by its key.
func (c *Client) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	resp, err := c.client.Object.Get(ctx, objectKey, nil)
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", objectKey, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get object %s: status %d", objectKey, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// PresignPutURL generates a presigned PUT URL using the COS SDK.
func (c *Client) PresignPutURL(ctx context.Context, bizType BizType, filename string, expire time.Duration) (uploadURL, objectKey, objectURL string, err error) {
	if !ValidBizTypes()[bizType] {
		return "", "", "", fmt.Errorf("invalid biz_type: %s", bizType)
	}

	objectKey = fmt.Sprintf("%s/%s_%s", bizType, uuid.New().String(), filename)

	presignedURL, err := c.client.Object.GetPresignedURL(ctx, http.MethodPut, objectKey, c.secretID, c.secretKey, expire, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("generate presigned url: %w", err)
	}

	return presignedURL.String(), objectKey, c.baseURL + "/" + objectKey, nil
}

// ObjectURL returns the full COS URL for an object key.
func (c *Client) ObjectURL(objectKey string) string {
	return c.baseURL + "/" + objectKey
}

// IsFullURL checks if a URL is already a full COS URL.
func IsFullURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")
}
