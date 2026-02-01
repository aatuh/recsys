package jsonl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

type S3RawEventSource struct {
	client *minio.Client
	bucket string
	prefix string
}

var _ datasource.RawEventSource = (*S3RawEventSource)(nil)

type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
	Prefix    string
	UseSSL    bool
}

func New(cfg Config) (*S3RawEventSource, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("s3 endpoint is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 bucket is required")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &S3RawEventSource{
		client: client,
		bucket: cfg.Bucket,
		prefix: strings.TrimPrefix(cfg.Prefix, "/"),
	}, nil
}

func (s *S3RawEventSource) ReadExposureEvents(
	ctx context.Context,
	tenant string,
	surface string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 256)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		if s == nil || s.client == nil {
			errs <- fmt.Errorf("s3 source not configured")
			return
		}

		keys := s.resolveKeys(ctx, tenant, surface, w)
		if len(keys) == 0 {
			return
		}
		for _, key := range keys {
			if err := s.readObject(ctx, key, w, out); err != nil {
				errs <- err
				return
			}
		}
	}()

	return out, errs
}

func (s *S3RawEventSource) resolveKeys(ctx context.Context, tenant, surface string, w windows.Window) []string {
	flatKey := s.objectKey("exposure.jsonl")
	if s.objectExists(ctx, flatKey) {
		return []string{flatKey}
	}

	var keys []string
	startDay := time.Date(w.Start.Year(), w.Start.Month(), w.Start.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(w.End.Year(), w.End.Month(), w.End.Day(), 0, 0, 0, 0, time.UTC)
	for day := startDay; day.Before(endDay); day = day.Add(24 * time.Hour) {
		name := fmt.Sprintf("exposure.%s.jsonl", day.Format("2006-01-02"))
		key := s.objectKey(path.Join(tenant, surface, name))
		if s.objectExists(ctx, key) {
			keys = append(keys, key)
		}
	}
	return keys
}

func (s *S3RawEventSource) objectKey(suffix string) string {
	if s.prefix == "" {
		return suffix
	}
	return path.Join(s.prefix, suffix)
}

func (s *S3RawEventSource) objectExists(ctx context.Context, key string) bool {
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	return err == nil
}

func (s *S3RawEventSource) readObject(
	ctx context.Context,
	key string,
	w windows.Window,
	out chan<- events.ExposureEvent,
) error {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, obj); err != nil {
		return err
	}
	scanner := bufio.NewScanner(buf)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var e events.ExposureEvent
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return fmt.Errorf("decode exposure jsonl: %w", err)
		}
		e = e.Normalized()
		if err := e.Validate(); err != nil {
			return fmt.Errorf("invalid exposure event: %w", err)
		}
		if w.Contains(e.TS.UTC()) {
			out <- e
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
