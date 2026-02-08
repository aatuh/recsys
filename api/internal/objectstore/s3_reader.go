package objectstore

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config configures an S3-compatible reader.
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}

// S3Reader reads objects from S3-compatible stores.
type S3Reader struct {
	client   *minio.Client
	maxBytes int64
}

func NewS3Reader(cfg S3Config, maxBytes int) (*S3Reader, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("s3 endpoint is required")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Region: cfg.Region,
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &S3Reader{client: client, maxBytes: int64(maxBytes)}, nil
}

func (r *S3Reader) Get(ctx context.Context, uri string) (_ []byte, err error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("s3 reader not configured")
	}
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return nil, err
	}
	obj, err := r.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := obj.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	var reader io.Reader = obj
	if r.maxBytes > 0 {
		reader = io.LimitReader(obj, r.maxBytes+1)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		if minioErr, ok := err.(minio.ErrorResponse); ok && minioErr.Code == "NoSuchKey" {
			return nil, ErrNotFound{URI: uri}
		}
		return nil, err
	}
	if r.maxBytes > 0 && int64(len(data)) > r.maxBytes {
		return nil, fmt.Errorf("s3 object too large")
	}
	return data, nil
}

func parseS3URI(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("s3 uri is empty")
	}
	if !strings.HasPrefix(raw, "s3://") {
		return "", "", fmt.Errorf("unsupported s3 uri: %s", raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	bucket := u.Host
	key := strings.TrimPrefix(u.Path, "/")
	if bucket == "" || key == "" {
		return "", "", fmt.Errorf("invalid s3 uri: %s", raw)
	}
	return bucket, key, nil
}
