package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type EnvConfig struct {
	OutDir         string            `json:"out_dir"`
	RawEventsDir   string            `json:"raw_events_dir"`
	CanonicalDir   string            `json:"canonical_dir"`
	CheckpointDir  string            `json:"checkpoint_dir"`
	RawSource      RawSourceConfig   `json:"raw_source"`
	ArtifactsDir   string            `json:"artifacts_dir"`
	ObjectStoreDir string            `json:"object_store_dir"`
	ObjectStore    ObjectStoreConfig `json:"object_store"`
	RegistryDir    string            `json:"registry_dir"`
	DB             DatabaseConfig    `json:"db"`
	Limits         Limits            `json:"limits"`
}

type Limits struct {
	MaxDaysBackfill        int `json:"max_days_backfill"`
	MaxEventsPerRun        int `json:"max_events_per_run"`
	MaxSessionsPerRun      int `json:"max_sessions_per_run"`
	MaxItemsPerSession     int `json:"max_items_per_session"`
	MaxDistinctItemsPerRun int `json:"max_distinct_items_per_run"`
	MaxNeighborsPerItem    int `json:"max_neighbors_per_item"`
	MaxItemsPerArtifact    int `json:"max_items_per_artifact"`
	MinCoocSupport         int `json:"min_cooc_support"`
}

type ObjectStoreConfig struct {
	Type string   `json:"type"`
	Dir  string   `json:"dir,omitempty"`
	S3   S3Config `json:"s3,omitempty"`
}

type S3Config struct {
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	UseSSL    bool   `json:"use_ssl,omitempty"`
}

type RawSourceConfig struct {
	Type     string               `json:"type"`
	Dir      string               `json:"dir,omitempty"`
	S3       S3Config             `json:"s3,omitempty"`
	Postgres PostgresSourceConfig `json:"postgres,omitempty"`
	Kafka    KafkaSourceConfig    `json:"kafka,omitempty"`
}

type PostgresSourceConfig struct {
	DSN           string `json:"dsn,omitempty"`
	TenantTable   string `json:"tenant_table,omitempty"`
	ExposureTable string `json:"exposure_table,omitempty"`
}

type KafkaSourceConfig struct {
	Brokers []string `json:"brokers,omitempty"`
	Topic   string   `json:"topic,omitempty"`
	GroupID string   `json:"group_id,omitempty"`
}

type DatabaseConfig struct {
	DSN               string `json:"dsn,omitempty"`
	AutoCreateTenant  bool   `json:"auto_create_tenant,omitempty"`
	StatementTimeoutS int    `json:"statement_timeout_s,omitempty"`
}

func LoadEnvConfig(path string) (EnvConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return EnvConfig{}, err
	}
	var c EnvConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return EnvConfig{}, fmt.Errorf("parse config: %w", err)
	}
	if c.OutDir == "" {
		c.OutDir = ".out"
	}
	if c.ObjectStore.Type == "" {
		if c.ObjectStoreDir != "" {
			c.ObjectStore.Type = "fs"
			c.ObjectStore.Dir = c.ObjectStoreDir
		} else {
			c.ObjectStore.Type = "fs"
			c.ObjectStore.Dir = ".out/objectstore"
		}
	}
	if c.ObjectStore.Type == "fs" && c.ObjectStore.Dir == "" {
		if c.ObjectStoreDir != "" {
			c.ObjectStore.Dir = c.ObjectStoreDir
		} else {
			c.ObjectStore.Dir = ".out/objectstore"
		}
	}
	if c.ObjectStoreDir == "" {
		c.ObjectStoreDir = c.ObjectStore.Dir
	}
	if c.CheckpointDir == "" {
		c.CheckpointDir = ".out/checkpoints"
	}
	if c.RawSource.Type == "" {
		c.RawSource.Type = "fs"
	}
	if c.RawSource.Type == "fs" && c.RawSource.Dir == "" {
		if c.RawEventsDir != "" {
			c.RawSource.Dir = c.RawEventsDir
		} else {
			c.RawSource.Dir = ".out/raw"
		}
	}
	if c.RawEventsDir == "" {
		c.RawEventsDir = c.RawSource.Dir
	}
	if c.RawSource.Postgres.TenantTable == "" {
		c.RawSource.Postgres.TenantTable = "tenants"
	}
	if c.RawSource.Postgres.ExposureTable == "" {
		c.RawSource.Postgres.ExposureTable = "exposure_events"
	}
	if c.Limits.MaxDaysBackfill == 0 {
		c.Limits.MaxDaysBackfill = 365
	}
	if c.Limits.MaxEventsPerRun == 0 {
		c.Limits.MaxEventsPerRun = 1_000_000
	}
	if c.Limits.MaxNeighborsPerItem == 0 {
		c.Limits.MaxNeighborsPerItem = 50
	}
	if c.Limits.MaxItemsPerArtifact == 0 {
		c.Limits.MaxItemsPerArtifact = 1000
	}
	if c.Limits.MaxSessionsPerRun == 0 {
		c.Limits.MaxSessionsPerRun = 1_000_000
	}
	if c.Limits.MaxItemsPerSession == 0 {
		c.Limits.MaxItemsPerSession = 200
	}
	if c.Limits.MaxDistinctItemsPerRun == 0 {
		c.Limits.MaxDistinctItemsPerRun = 2_000_000
	}
	if c.Limits.MinCoocSupport == 0 {
		c.Limits.MinCoocSupport = 2
	}
	return c, nil
}
