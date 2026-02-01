package factory

import (
	"fmt"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/objectstore/fs"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/objectstore/s3"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/objectstore"
)

// BuildObjectStore constructs an object store adapter from config.
func BuildObjectStore(cfg config.EnvConfig) (objectstore.ObjectStore, error) {
	storeCfg := cfg.ObjectStore
	switch strings.ToLower(strings.TrimSpace(storeCfg.Type)) {
	case "", "fs", "file":
		return fs.New(storeCfg.Dir), nil
	case "s3", "minio":
		return s3.New(s3.Config{
			Endpoint:  storeCfg.S3.Endpoint,
			Bucket:    storeCfg.S3.Bucket,
			AccessKey: storeCfg.S3.AccessKey,
			SecretKey: storeCfg.S3.SecretKey,
			Region:    storeCfg.S3.Region,
			Prefix:    storeCfg.S3.Prefix,
			UseSSL:    storeCfg.S3.UseSSL,
		})
	default:
		return nil, fmt.Errorf("unsupported object store type: %s", storeCfg.Type)
	}
}
