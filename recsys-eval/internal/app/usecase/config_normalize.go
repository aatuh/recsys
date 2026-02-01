package usecase

import (
	"fmt"
	"strings"
)

// NormalizeAndValidate applies defaults and validates config for the selected mode.
func NormalizeAndValidate(dataset *DatasetConfig, eval *EvalConfig, mode string) error {
	normalizeEval(eval)
	normalizeDataset(dataset)

	mode = strings.ToLower(mode)
	if mode == "" {
		mode = strings.ToLower(eval.Mode)
	}
	if mode == "" {
		return fmt.Errorf("mode is required")
	}

	if err := validateDataset(*dataset, mode); err != nil {
		return err
	}
	if err := validateEval(*eval, mode); err != nil {
		return err
	}
	if err := validateScale(*dataset, *eval, mode); err != nil {
		return err
	}
	return nil
}

func normalizeDataset(_ *DatasetConfig) {
	// No defaults yet.
}

func normalizeEval(eval *EvalConfig) {
	if eval.Scale.Mode == "" {
		eval.Scale.Mode = "memory"
	}

	if eval.Offline.Bootstrap != nil {
		if eval.Offline.Bootstrap.Iterations == 0 {
			eval.Offline.Bootstrap.Iterations = 200
		}
		if eval.Offline.Bootstrap.CILevel == 0 {
			eval.Offline.Bootstrap.CILevel = 0.95
		}
	}

	if eval.Experiment.MultipleComparisons != nil {
		if eval.Experiment.MultipleComparisons.Method == "" {
			eval.Experiment.MultipleComparisons.Method = "fdr_bh"
		}
		if eval.Experiment.MultipleComparisons.Alpha == 0 {
			eval.Experiment.MultipleComparisons.Alpha = 0.05
		}
	}

	if eval.Experiment.SRM != nil {
		if eval.Experiment.SRM.Alpha == 0 {
			eval.Experiment.SRM.Alpha = 0.001
		}
		if eval.Experiment.SRM.MinSample == 0 {
			eval.Experiment.SRM.MinSample = 100
		}
	}

	if eval.Experiment.Decision != nil {
		if eval.Experiment.Decision.PrimaryThresholds == nil {
			eval.Experiment.Decision.PrimaryThresholds = map[string]float64{}
		}
	}

	if eval.Experiment.AACheck != nil && len(eval.Experiment.AACheck.Thresholds) == 0 {
		eval.Experiment.AACheck.Thresholds = []float64{0.05, 0.01, 0.001}
	}

	if eval.OPE.RewardEvent == "" {
		eval.OPE.RewardEvent = "click"
	}
	if eval.OPE.Unit == "" {
		eval.OPE.Unit = "request"
	}
	if eval.OPE.RewardAggregation == "" {
		eval.OPE.RewardAggregation = "sum"
	}
	if eval.OPE.MinPropensity == 0 {
		eval.OPE.MinPropensity = 1e-6
	}

	if eval.Interleaving.Algorithm == "" {
		eval.Interleaving.Algorithm = "team_draft"
	}
	if eval.Interleaving.Seed == 0 {
		eval.Interleaving.Seed = 42
	}

	if eval.Artifacts.ManifestURI != "" && eval.Artifacts.ObjectStore.Type == "" {
		eval.Artifacts.ObjectStore.Type = "file"
	}
}

func validateDataset(cfg DatasetConfig, mode string) error {
	switch mode {
	case "offline", "experiment", "ope":
		if cfg.Exposures.Type == "" || cfg.Outcomes.Type == "" {
			return fmt.Errorf("exposures and outcomes sources are required")
		}
		if err := validateSource(cfg.Exposures, "exposures"); err != nil {
			return err
		}
		if err := validateSource(cfg.Outcomes, "outcomes"); err != nil {
			return err
		}
	case "interleaving":
		if cfg.Interleaving == nil {
			return fmt.Errorf("interleaving dataset config is required")
		}
		if err := validateSource(cfg.Interleaving.RankerA, "ranker_a"); err != nil {
			return err
		}
		if err := validateSource(cfg.Interleaving.RankerB, "ranker_b"); err != nil {
			return err
		}
		if err := validateSource(cfg.Interleaving.Outcomes, "outcomes"); err != nil {
			return err
		}
	case "aa-check":
		if cfg.Assignments.Type == "" {
			return fmt.Errorf("assignments source is required for aa-check")
		}
		if err := validateSource(cfg.Assignments, "assignments"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
	return nil
}

func validateEval(cfg EvalConfig, mode string) error {
	switch mode {
	case "offline":
		if len(cfg.Offline.Metrics) == 0 {
			return fmt.Errorf("offline.metrics is required")
		}
	case "experiment":
		// guardrails can be zeroed; no required fields.
	case "ope":
		if cfg.OPE.RewardEvent != "click" && cfg.OPE.RewardEvent != "conversion" {
			return fmt.Errorf("ope.reward_event must be click or conversion")
		}
		if cfg.OPE.Unit != "item" && cfg.OPE.Unit != "request" {
			return fmt.Errorf("ope.unit must be item or request")
		}
		if cfg.OPE.RewardAggregation != "sum" && cfg.OPE.RewardAggregation != "max" {
			return fmt.Errorf("ope.reward_aggregation must be sum or max")
		}
	case "interleaving":
		if cfg.Interleaving.Algorithm == "" {
			return fmt.Errorf("interleaving.algorithm is required")
		}
	case "aa-check":
		if cfg.Experiment.AACheck == nil {
			return fmt.Errorf("experiment.aa_check config is required")
		}
		if !cfg.Experiment.AACheck.Enabled {
			return fmt.Errorf("experiment.aa_check.enabled must be true for aa-check mode")
		}
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
	if cfg.Artifacts.ManifestURI != "" {
		storeType := strings.ToLower(strings.TrimSpace(cfg.Artifacts.ObjectStore.Type))
		if storeType == "" {
			storeType = "file"
		}
		if storeType == "s3" || storeType == "minio" {
			s3 := cfg.Artifacts.ObjectStore.S3
			if s3.Endpoint == "" || s3.Bucket == "" {
				return fmt.Errorf("artifacts.object_store.s3 endpoint and bucket are required")
			}
		}
	}
	return nil
}

func validateScale(dataset DatasetConfig, eval EvalConfig, mode string) error {
	scale := strings.ToLower(eval.Scale.Mode)
	if scale == "" {
		scale = "memory"
	}
	switch scale {
	case "memory":
		return nil
	case "stream":
		if dataset.Exposures.Type != "" && !strings.EqualFold(dataset.Exposures.Type, "jsonl") {
			return fmt.Errorf("stream mode requires jsonl exposures")
		}
		if dataset.Outcomes.Type != "" && !strings.EqualFold(dataset.Outcomes.Type, "jsonl") {
			return fmt.Errorf("stream mode requires jsonl outcomes")
		}
		if mode == "offline" {
			if eval.Offline.TimeSplit != nil {
				return fmt.Errorf("offline.time_split is not supported in stream mode")
			}
			if eval.Offline.Bootstrap != nil && eval.Offline.Bootstrap.Enabled {
				return fmt.Errorf("offline.bootstrap is not supported in stream mode")
			}
		}
		if mode == "ope" {
			return fmt.Errorf("ope stream mode is not supported")
		}
		return nil
	case "duckdb":
		if !duckdbSupported() {
			return fmt.Errorf("duckdb mode is not supported in this build (use -tags duckdb)")
		}
		if dataset.Exposures.Type != "" && !strings.EqualFold(dataset.Exposures.Type, "duckdb") {
			return fmt.Errorf("duckdb mode requires duckdb exposures")
		}
		if dataset.Outcomes.Type != "" && !strings.EqualFold(dataset.Outcomes.Type, "duckdb") {
			return fmt.Errorf("duckdb mode requires duckdb outcomes")
		}
		if dataset.Assignments.Type != "" && !strings.EqualFold(dataset.Assignments.Type, "duckdb") {
			return fmt.Errorf("duckdb mode requires duckdb assignments")
		}
		return nil
	default:
		return fmt.Errorf("scale.mode must be memory, stream, or duckdb")
	}
}

func validateSource(src SourceConfig, name string) error {
	switch strings.ToLower(src.Type) {
	case "jsonl":
		if src.Path == "" {
			return fmt.Errorf("%s.path is required for jsonl source", name)
		}
	case "postgres", "duckdb":
		if src.Query == "" {
			return fmt.Errorf("%s.query is required for %s source", name, src.Type)
		}
		if src.DSN == "" {
			return fmt.Errorf("%s.dsn is required for %s source", name, src.Type)
		}
	default:
		return fmt.Errorf("unsupported %s source type: %s", name, src.Type)
	}
	return nil
}
