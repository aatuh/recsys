package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/objectstore"
)

// Config configures an S3-compatible object store.
type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
	Prefix    string
	UseSSL    bool
}

// Store implements objectstore.ObjectStore using S3-compatible APIs.
type Store struct {
	client *minio.Client
	bucket string
	prefix string
	ensure sync.Once
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
	return &Store{
		client: client,
		bucket: cfg.Bucket,
		prefix: strings.Trim(strings.TrimSpace(cfg.Prefix), "/"),
	}, nil
}

func (s *Store) Put(ctx context.Context, key string, contentType string, data []byte) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("s3 store not configured")
	}
	if err := s.ensureBucket(ctx); err != nil {
		return "", err
	}
	objKey := s.normalizeKey(key)
	_, err := s.client.PutObject(ctx, s.bucket, objKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, objKey), nil
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

func (s *Store) ensureBucket(ctx context.Context) error {
	var err error
	s.ensure.Do(func() {
		exists, e := s.client.BucketExists(ctx, s.bucket)
		if e != nil {
			err = e
			return
		}
		if exists {
			return
		}
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	})
	return err
}

func (s *Store) normalizeKey(key string) string {
	clean := strings.TrimPrefix(key, "/")
	if s.prefix == "" {
		return clean
	}
	return path.Join(s.prefix, clean)
}

func parseS3URI(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("s3 uri is empty")
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

var _ objectstore.ObjectStore = (*Store)(nil)
