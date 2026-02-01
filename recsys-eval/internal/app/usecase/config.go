package usecase

import "github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"

// DatasetConfig describes where to read exposures/outcomes/assignments.
type DatasetConfig struct {
	Exposures    SourceConfig               `yaml:"exposures"`
	Outcomes     SourceConfig               `yaml:"outcomes"`
	Assignments  SourceConfig               `yaml:"assignments,omitempty"`
	Interleaving *InterleavingDatasetConfig `yaml:"interleaving,omitempty"`
}

// SourceConfig supports JSONL, Postgres, and (optional) DuckDB queries that return JSON rows.
type SourceConfig struct {
	Type  string `yaml:"type"`            // jsonl | postgres | duckdb
	Path  string `yaml:"path,omitempty"`  // for jsonl
	DSN   string `yaml:"dsn,omitempty"`   // for postgres
	Query string `yaml:"query,omitempty"` // for postgres
}

// EvalConfig is a top-level config wrapper.
type EvalConfig struct {
	Mode         string             `yaml:"mode"` // offline | experiment | ope | interleaving | aa-check
	Offline      OfflineConfig      `yaml:"offline"`
	Experiment   ExperimentConfig   `yaml:"experiment"`
	OPE          OPEConfig          `yaml:"ope"`
	Interleaving InterleavingConfig `yaml:"interleaving"`
	Scale        ScaleConfig        `yaml:"scale"`
	Artifacts    ArtifactConfig     `yaml:"artifacts,omitempty"`
}

// ArtifactConfig describes how to resolve artifact manifests.
type ArtifactConfig struct {
	ManifestURI string            `yaml:"manifest_uri,omitempty"`
	ObjectStore ObjectStoreConfig `yaml:"object_store,omitempty"`
}

// ObjectStoreConfig configures artifact object storage.
type ObjectStoreConfig struct {
	Type string   `yaml:"type,omitempty"` // file | s3 | minio
	S3   S3Config `yaml:"s3,omitempty"`
}

// S3Config configures S3/MinIO access.
type S3Config struct {
	Endpoint  string `yaml:"endpoint,omitempty"`
	Bucket    string `yaml:"bucket,omitempty"`
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Region    string `yaml:"region,omitempty"`
	UseSSL    bool   `yaml:"use_ssl,omitempty"`
}

// OfflineConfig controls offline evaluation.
type OfflineConfig struct {
	Metrics    []metrics.MetricSpec `yaml:"metrics"`
	SliceKeys  []string             `yaml:"slice_keys"`
	Gates      []GateSpec           `yaml:"gates"`
	TimeSplit  *TimeSplitConfig     `yaml:"time_split,omitempty"`
	Bootstrap  *BootstrapConfig     `yaml:"bootstrap,omitempty"`
	HistoryDir string               `yaml:"history_dir,omitempty"`
}

// ExperimentConfig controls experiment analysis and guardrails.
type ExperimentConfig struct {
	ExperimentID        string                     `yaml:"experiment_id"`
	SliceKeys           []string                   `yaml:"slice_keys"`
	Guardrails          GuardrailConfig            `yaml:"guardrails"`
	ControlVariant      string                     `yaml:"control_variant,omitempty"`
	PrimaryMetrics      []string                   `yaml:"primary_metrics,omitempty"`
	CUPED               *CUPEDConfig               `yaml:"cuped,omitempty"`
	Power               *PowerConfig               `yaml:"power,omitempty"`
	MultipleComparisons *MultipleComparisonsConfig `yaml:"multiple_comparisons,omitempty"`
	SRM                 *SRMConfig                 `yaml:"srm,omitempty"`
	Decision            *DecisionConfig            `yaml:"decision,omitempty"`
	AACheck             *AACheckConfig             `yaml:"aa_check,omitempty"`
}

// GateSpec defines a regression gate.
type GateSpec struct {
	Metric  string  `yaml:"metric"`
	MaxDrop float64 `yaml:"max_drop"`
}

// GuardrailConfig sets absolute thresholds for experiment guardrails.
type GuardrailConfig struct {
	MaxLatencyP95Ms float64 `yaml:"max_latency_p95_ms"`
	MaxErrorRate    float64 `yaml:"max_error_rate"`
	MaxEmptyRate    float64 `yaml:"max_empty_rate"`
}

// OPEConfig controls counterfactual evaluation.
type OPEConfig struct {
	RewardEvent                  string  `yaml:"reward_event"`                 // click | conversion
	Unit                         string  `yaml:"unit,omitempty"`               // item | request
	RewardAggregation            string  `yaml:"reward_aggregation,omitempty"` // sum | max
	TopK                         int     `yaml:"top_k,omitempty"`
	Clipping                     float64 `yaml:"clipping,omitempty"`
	MinPropensity                float64 `yaml:"min_propensity,omitempty"`
	EnableSNIPS                  bool    `yaml:"enable_snips,omitempty"`
	EnableDR                     bool    `yaml:"enable_dr,omitempty"`
	PositionDiscount             string  `yaml:"position_discount,omitempty"` // none | log2
	AllowMissingTargetPropensity bool    `yaml:"allow_missing_target_propensity,omitempty"`
}

// InterleavingConfig controls interleaving analysis.
type InterleavingConfig struct {
	Algorithm  string `yaml:"algorithm"` // team_draft | balanced | optimized
	Seed       int64  `yaml:"seed,omitempty"`
	MaxResults int    `yaml:"max_results,omitempty"`
}

// ScaleConfig controls large-dataset evaluation modes.
type ScaleConfig struct {
	Mode   string       `yaml:"mode"` // memory | stream | duckdb
	Stream StreamConfig `yaml:"stream,omitempty"`
	DuckDB DuckDBConfig `yaml:"duckdb,omitempty"`
}

// StreamConfig controls streaming join behavior.
type StreamConfig struct {
	MaxOpenRequests int    `yaml:"max_open_requests,omitempty"`
	TempDir         string `yaml:"temp_dir,omitempty"`
}

// DuckDBConfig controls DuckDB pushdown mode.
type DuckDBConfig struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

// InterleavingDatasetConfig specifies input lists for interleaving.
type InterleavingDatasetConfig struct {
	RankerA  SourceConfig `yaml:"ranker_a"`
	RankerB  SourceConfig `yaml:"ranker_b"`
	Outcomes SourceConfig `yaml:"outcomes"`
}

// TimeSplitConfig enforces train/test windows for offline evaluation.
type TimeSplitConfig struct {
	Train TimeWindow `yaml:"train"`
	Test  TimeWindow `yaml:"test"`
}

// TimeWindow defines an inclusive [start, end] time range in RFC3339 format.
type TimeWindow struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

// BootstrapConfig controls confidence interval estimation.
type BootstrapConfig struct {
	Enabled    bool    `yaml:"enabled"`
	Iterations int     `yaml:"iterations"`
	Seed       int64   `yaml:"seed"`
	CILevel    float64 `yaml:"ci_level"`
}

// CUPEDConfig configures variance reduction for experiment analysis.
type CUPEDConfig struct {
	Enabled      bool   `yaml:"enabled"`
	CovariateKey string `yaml:"covariate_key"`
}

// PowerConfig controls sample size and runtime estimation.
type PowerConfig struct {
	Metric       string  `yaml:"metric"` // ctr | conversion_rate
	BaselineRate float64 `yaml:"baseline_rate"`
	MDE          float64 `yaml:"mde"`
	Alpha        float64 `yaml:"alpha"`
	Power        float64 `yaml:"power"`
	DailySamples int     `yaml:"daily_samples"`
}

// MultipleComparisonsConfig controls p-value correction.
type MultipleComparisonsConfig struct {
	Method string  `yaml:"method"` // fdr_bh | bonferroni | none
	Alpha  float64 `yaml:"alpha"`
}

// SRMConfig controls sample ratio mismatch detection.
type SRMConfig struct {
	Enabled     bool               `yaml:"enabled"`
	Alpha       float64            `yaml:"alpha,omitempty"`
	MinSample   int                `yaml:"min_sample,omitempty"`
	Expected    map[string]float64 `yaml:"expected,omitempty"`
	SliceKeys   []string           `yaml:"slice_keys,omitempty"`
	GateEnabled bool               `yaml:"gate_enabled,omitempty"`
}

// DecisionConfig controls decision artifact generation.
type DecisionConfig struct {
	Enabled           bool               `yaml:"enabled"`
	OutputPath        string             `yaml:"output_path,omitempty"`
	CandidateVariant  string             `yaml:"candidate_variant,omitempty"`
	PrimaryThresholds map[string]float64 `yaml:"primary_thresholds,omitempty"`
	GuardrailGate     bool               `yaml:"guardrail_gate,omitempty"`
	SRMGate           bool               `yaml:"srm_gate,omitempty"`
}

// AACheckConfig controls A/A sanity check behavior.
type AACheckConfig struct {
	Enabled    bool      `yaml:"enabled"`
	Thresholds []float64 `yaml:"thresholds,omitempty"`
}
