package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/objectstore"
)

// Config configures S3/MinIO access.
type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}

// Store reads artifacts from S3-compatible storage.
type Store struct {
	client *minio.Client
	bucket string
}

func New(cfg Config) (*Store, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("s3 endpoint is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("s3 bucket is required")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Region: cfg.Region,
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Store{client: client, bucket: cfg.Bucket}, nil
}

func (s *Store) Get(ctx context.Context, uri string) ([]byte, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("s3 store not configured")
	}
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return nil, err
	}
	if bucket == "" {
		bucket = s.bucket
	}
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func parseS3URI(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("empty s3 uri")
	}
	if strings.HasPrefix(raw, "s3://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", "", err
		}
		bucket := u.Host
		key := strings.TrimPrefix(u.Path, "/")
		if bucket == "" || key == "" {
			return "", "", fmt.Errorf("invalid s3 uri")
		}
		return bucket, key, nil
	}
	return "", raw, nil
}

var _ objectstore.Store = (*Store)(nil)
