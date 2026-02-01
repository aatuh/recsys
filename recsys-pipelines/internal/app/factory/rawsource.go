package factory

import (
	"fmt"
	"strings"

	fs "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/datasource/files/jsonl"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/datasource/kafka"
	pgsrc "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/datasource/postgres"
	s3jsonl "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/datasource/s3/jsonl"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

func BuildRawSource(cfg config.EnvConfig) (datasource.RawEventSource, func(), error) {
	typ := strings.ToLower(strings.TrimSpace(cfg.RawSource.Type))
	switch typ {
	case "", "fs", "files", "file":
		dir := cfg.RawSource.Dir
		if dir == "" {
			dir = cfg.RawEventsDir
		}
		return fs.New(dir), nil, nil
	case "s3", "minio":
		src, err := s3jsonl.New(s3jsonl.Config{
			Endpoint:  cfg.RawSource.S3.Endpoint,
			Bucket:    cfg.RawSource.S3.Bucket,
			AccessKey: cfg.RawSource.S3.AccessKey,
			SecretKey: cfg.RawSource.S3.SecretKey,
			Region:    cfg.RawSource.S3.Region,
			Prefix:    cfg.RawSource.S3.Prefix,
			UseSSL:    cfg.RawSource.S3.UseSSL,
		})
		if err != nil {
			return nil, nil, err
		}
		return src, nil, nil
	case "postgres", "postgresql":
		src, err := pgsrc.New(pgsrc.Config{
			DSN:           pick(cfg.RawSource.Postgres.DSN, cfg.DB.DSN),
			TenantTable:   cfg.RawSource.Postgres.TenantTable,
			ExposureTable: cfg.RawSource.Postgres.ExposureTable,
		})
		if err != nil {
			return nil, nil, err
		}
		return src, src.Close, nil
	case "kafka":
		src := kafka.New(cfg.RawSource.Kafka.Brokers, cfg.RawSource.Kafka.Topic, cfg.RawSource.Kafka.GroupID)
		return src, nil, nil
	default:
		return nil, nil, fmt.Errorf("unknown raw source type: %s", typ)
	}
}

func pick(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}
