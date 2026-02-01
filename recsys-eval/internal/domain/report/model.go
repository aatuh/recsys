package report

import (
	"encoding/json"
	"time"
)

// Report is the top-level evaluation artifact.
type Report struct {
	RunID                   string              `json:"run_id"`
	Mode                    string              `json:"mode"`
	CreatedAt               time.Time           `json:"created_at"`
	Version                 string              `json:"version"`
	BinaryVersion           string              `json:"binary_version,omitempty"`
	GitCommit               string              `json:"git_commit,omitempty"`
	EffectiveConfig         json.RawMessage     `json:"effective_config,omitempty"`
	InputDatasetFingerprint string              `json:"input_dataset_fingerprint,omitempty"`
	Artifacts               *ArtifactProvenance `json:"artifacts,omitempty"`
	Summary                 Summary             `json:"summary"`
	Offline                 *OfflineReport      `json:"offline,omitempty"`
	Experiment              *ExperimentReport   `json:"experiment,omitempty"`
	OPE                     *OPEReport          `json:"ope,omitempty"`
	Interleaving            *InterleavingReport `json:"interleaving,omitempty"`
	AA                      *AAReport           `json:"aa_check,omitempty"`
	DataQuality             *DataQualityReport  `json:"data_quality,omitempty"`
	Gates                   []GateResult        `json:"gates,omitempty"`
	Warnings                []string            `json:"warnings,omitempty"`
}

const (
	ModeOffline      = "offline"
	ModeExperiment   = "experiment"
	ModeOPE          = "ope"
	ModeInterleaving = "interleaving"
	ModeAACheck      = "aa-check"
)

// SupportedModes lists report mode values.
func SupportedModes() []string {
	return []string{ModeOffline, ModeExperiment, ModeOPE, ModeInterleaving, ModeAACheck}
}

// Summary captures high-level run info.
type Summary struct {
	CasesEvaluated int `json:"cases_evaluated"`
}

// ArtifactProvenance captures manifest and artifact metadata for the run.
type ArtifactProvenance struct {
	ManifestURI string        `json:"manifest_uri,omitempty"`
	Tenant      string        `json:"tenant,omitempty"`
	Surface     string        `json:"surface,omitempty"`
	UpdatedAt   string        `json:"updated_at,omitempty"`
	Artifacts   []ArtifactRef `json:"artifacts,omitempty"`
	Warnings    []string      `json:"warnings,omitempty"`
}

// ArtifactRef captures a resolved artifact from the manifest.
type ArtifactRef struct {
	Type       string `json:"type,omitempty"`
	URI        string `json:"uri,omitempty"`
	Version    string `json:"version,omitempty"`
	SourceHash string `json:"source_hash,omitempty"`
	BuiltAt    string `json:"built_at,omitempty"`
	Checksum   string `json:"checksum,omitempty"`
}

// MetricResult stores a single metric value.
type MetricResult struct {
	Name  string    `json:"name"`
	Value float64   `json:"value"`
	CI    *MetricCI `json:"ci,omitempty"`
}

// OfflineReport captures offline metric results.
type OfflineReport struct {
	Metrics   []MetricResult            `json:"metrics"`
	BySegment map[string][]MetricResult `json:"by_segment,omitempty"`
	Baseline  *BaselineComparison       `json:"baseline,omitempty"`
}

// BaselineComparison stores deltas vs baseline.
type BaselineComparison struct {
	BaselineRunID string        `json:"baseline_run_id"`
	Deltas        []MetricDelta `json:"deltas"`
}

// MetricDelta captures metric changes vs baseline.
type MetricDelta struct {
	Name          string  `json:"name"`
	Delta         float64 `json:"delta"`
	RelativeDelta float64 `json:"relative_delta,omitempty"`
	Before        float64 `json:"before"`
	After         float64 `json:"after"`
}

// GateResult stores regression gate results.
type GateResult struct {
	Metric  string  `json:"metric"`
	MaxDrop float64 `json:"max_drop"`
	Delta   float64 `json:"delta"`
	Passed  bool    `json:"passed"`
}

// ExperimentReport captures online experiment results.
type ExperimentReport struct {
	ExperimentID string                      `json:"experiment_id"`
	Variants     []VariantMetrics            `json:"variants"`
	BySegment    map[string][]VariantMetrics `json:"by_segment,omitempty"`
	Guardrails   GuardrailReport             `json:"guardrails"`
	Cuped        *CupedReport                `json:"cuped,omitempty"`
	Power        *PowerAnalysisReport        `json:"power,omitempty"`
	Significance []ExperimentSignificance    `json:"significance,omitempty"`
	SRM          *SRMReport                  `json:"srm,omitempty"`
}

// VariantMetrics holds metrics for a variant.
type VariantMetrics struct {
	Variant                string   `json:"variant"`
	Exposures              int      `json:"exposures"`
	Clicks                 int      `json:"clicks"`
	Conversions            int      `json:"conversions"`
	CTR                    float64  `json:"ctr"`
	ConversionRate         float64  `json:"conversion_rate"`
	Revenue                float64  `json:"revenue"`
	RevenuePerRequest      float64  `json:"revenue_per_request"`
	CTRAdjusted            *float64 `json:"ctr_adjusted,omitempty"`
	ConversionRateAdjusted *float64 `json:"conversion_rate_adjusted,omitempty"`
}

// GuardrailReport summarizes guardrail metrics.
type GuardrailReport struct {
	PerVariant []GuardrailMetrics `json:"per_variant"`
}

// GuardrailMetrics holds guardrail values for a variant.
type GuardrailMetrics struct {
	Variant       string  `json:"variant"`
	LatencyP95Ms  float64 `json:"latency_p95_ms"`
	ErrorRate     float64 `json:"error_rate"`
	EmptyRate     float64 `json:"empty_rate"`
	PassLatency   bool    `json:"pass_latency"`
	PassErrorRate bool    `json:"pass_error_rate"`
	PassEmptyRate bool    `json:"pass_empty_rate"`
}

// MetricCI captures confidence interval for a metric value.
type MetricCI struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Level float64 `json:"level"`
}

// CupedReport summarizes CUPED variance reduction.
type CupedReport struct {
	Enabled      bool    `json:"enabled"`
	CovariateKey string  `json:"covariate_key,omitempty"`
	ThetaCTR     float64 `json:"theta_ctr,omitempty"`
	ThetaCVR     float64 `json:"theta_conversion_rate,omitempty"`
	CoverageRate float64 `json:"coverage_rate,omitempty"`
}

// PowerAnalysisReport summarizes sample size and runtime estimation.
type PowerAnalysisReport struct {
	Metric               string  `json:"metric"`
	BaselineRate         float64 `json:"baseline_rate"`
	MDE                  float64 `json:"mde"`
	Alpha                float64 `json:"alpha"`
	Power                float64 `json:"power"`
	SampleSizePerVariant int     `json:"sample_size_per_variant"`
	EstimatedDays        float64 `json:"estimated_days,omitempty"`
}

// ExperimentSignificance reports corrected significance per metric.
type ExperimentSignificance struct {
	Metric         string  `json:"metric"`
	Variant        string  `json:"variant"`
	Segment        string  `json:"segment,omitempty"`
	PValue         float64 `json:"p_value"`
	AdjustedPValue float64 `json:"adjusted_p_value,omitempty"`
	Significant    bool    `json:"significant"`
	Primary        bool    `json:"primary"`
	Correction     string  `json:"correction,omitempty"`
}

// OPEReport captures off-policy evaluation results.
type OPEReport struct {
	Unit              string               `json:"unit"`
	RewardAggregation string               `json:"reward_aggregation"`
	Estimators        []OPEEstimatorResult `json:"estimators"`
	WeightStats       WeightStats          `json:"weight_stats"`
	Diagnostics       OPEDiagnostics       `json:"diagnostics"`
}

// OPEEstimatorResult stores estimator outputs.
type OPEEstimatorResult struct {
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	Variance float64 `json:"variance,omitempty"`
}

// WeightStats captures importance weight diagnostics.
type WeightStats struct {
	Count               int     `json:"count"`
	Mean                float64 `json:"mean"`
	P95                 float64 `json:"p95"`
	Max                 float64 `json:"max"`
	EffectiveSampleSize float64 `json:"effective_sample_size"`
	ClippedFraction     float64 `json:"clipped_fraction"`
}

// OPEDiagnostics captures propensity and coverage diagnostics.
type OPEDiagnostics struct {
	MissingTargetPropensity int `json:"missing_target_propensity"`
	NearZeroPropensity      int `json:"near_zero_propensity"`
	TotalItems              int `json:"total_items"`
}

// InterleavingReport captures interleaving analysis results.
type InterleavingReport struct {
	Algorithm string  `json:"algorithm"`
	RankerA   string  `json:"ranker_a"`
	RankerB   string  `json:"ranker_b"`
	Requests  int     `json:"requests"`
	WinsA     int     `json:"wins_a"`
	WinsB     int     `json:"wins_b"`
	Ties      int     `json:"ties"`
	WinRateA  float64 `json:"win_rate_a"`
	WinRateB  float64 `json:"win_rate_b"`
	PValue    float64 `json:"p_value,omitempty"`
}

// AAReport summarizes A/A sanity check results.
type AAReport struct {
	Metrics    []AAPValueSummary `json:"metrics"`
	Thresholds []float64         `json:"thresholds"`
	TotalTests int               `json:"total_tests"`
	Uniformish bool              `json:"uniformish"`
}

// AAPValueSummary captures p-value distributions.
type AAPValueSummary struct {
	Metric    string         `json:"metric"`
	BySegment bool           `json:"by_segment"`
	Counts    map[string]int `json:"counts"`
}

// SRMReport captures sample ratio mismatch diagnostics.
type SRMReport struct {
	Alpha     float64              `json:"alpha"`
	MinSample int                  `json:"min_sample"`
	Global    SRMResult            `json:"global"`
	BySegment map[string]SRMResult `json:"by_segment,omitempty"`
}

// SRMResult is a chi-square test output.
type SRMResult struct {
	PValue     float64            `json:"p_value"`
	Observed   map[string]int     `json:"observed"`
	Expected   map[string]float64 `json:"expected"`
	Detected   bool               `json:"detected"`
	SampleSize int                `json:"sample_size"`
}

// DataQualityReport captures input validation statistics.
type DataQualityReport struct {
	ExposureCompleteness   CompletenessResult  `json:"exposure_completeness"`
	OutcomeCompleteness    CompletenessResult  `json:"outcome_completeness"`
	AssignmentCompleteness *CompletenessResult `json:"assignment_completeness,omitempty"`
	JoinIntegrity          JoinIntegrity       `json:"join_integrity"`
}

// CompletenessResult captures missing field rates.
type CompletenessResult struct {
	Total       int     `json:"total"`
	Missing     int     `json:"missing"`
	MissingRate float64 `json:"missing_rate"`
}

// JoinIntegrity captures join coverage.
type JoinIntegrity struct {
	ExposureCount      int     `json:"exposure_count"`
	OutcomeCount       int     `json:"outcome_count"`
	AssignmentsCount   int     `json:"assignments_count"`
	ExposureJoinRate   float64 `json:"exposure_join_rate"`
	OutcomeJoinRate    float64 `json:"outcome_join_rate"`
	AssignmentJoinRate float64 `json:"assignment_join_rate"`
}
