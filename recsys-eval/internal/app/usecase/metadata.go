package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/version"
)

// ReportMetadata carries build and config information for reports.
type ReportMetadata struct {
	BinaryVersion           string
	GitCommit               string
	EffectiveConfig         json.RawMessage
	InputDatasetFingerprint string
	Artifacts               *report.ArtifactProvenance
}

// BuildReportMetadata constructs report metadata with effective config and dataset fingerprint.
func BuildReportMetadata(dataset DatasetConfig, eval EvalConfig, mode string) (ReportMetadata, error) {
	cfgBytes, err := json.Marshal(struct {
		Mode    string        `json:"mode"`
		Dataset DatasetConfig `json:"dataset"`
		Eval    EvalConfig    `json:"eval"`
	}{Mode: mode, Dataset: dataset, Eval: eval})
	if err != nil {
		return ReportMetadata{}, err
	}
	fingerprint, err := FingerprintDataset(dataset, mode)
	if err != nil {
		return ReportMetadata{}, err
	}
	meta := ReportMetadata{
		BinaryVersion:           version.Version,
		GitCommit:               version.GitCommit,
		EffectiveConfig:         cfgBytes,
		InputDatasetFingerprint: fingerprint,
	}
	if eval.Artifacts.ManifestURI != "" {
		prov, err := ResolveArtifactProvenance(context.Background(), eval.Artifacts)
		if err != nil {
			if !errors.Is(err, ErrArtifactProvenanceDisabled) {
				return ReportMetadata{}, err
			}
		}
		meta.Artifacts = prov
	}
	return meta, nil
}

// FingerprintDataset computes a reproducible fingerprint for input sources.
func FingerprintDataset(dataset DatasetConfig, mode string) (string, error) {
	parts := []string{strings.ToLower(mode)}
	appendSource := func(name string, src SourceConfig) error {
		parts = append(parts, name, strings.ToLower(src.Type))
		switch strings.ToLower(src.Type) {
		case "jsonl", "csv", "parquet":
			if src.Path == "" {
				return fmt.Errorf("%s path is required", name)
			}
			info, err := os.Stat(src.Path)
			if err != nil {
				return err
			}
			parts = append(parts, filepath.Base(src.Path), fmt.Sprintf("%d", info.Size()))
			if info.Size() <= 10*1024*1024 {
				h, err := hashFile(src.Path)
				if err != nil {
					return err
				}
				parts = append(parts, h)
			}
		case "postgres", "duckdb":
			parts = append(parts, src.DSN, src.Query)
		default:
			return nil
		}
		return nil
	}

	if dataset.Interleaving != nil && mode == "interleaving" {
		if err := appendSource("ranker_a", dataset.Interleaving.RankerA); err != nil {
			return "", err
		}
		if err := appendSource("ranker_b", dataset.Interleaving.RankerB); err != nil {
			return "", err
		}
		if err := appendSource("outcomes", dataset.Interleaving.Outcomes); err != nil {
			return "", err
		}
	} else {
		if dataset.Exposures.Type != "" {
			if err := appendSource("exposures", dataset.Exposures); err != nil {
				return "", err
			}
		}
		if dataset.Outcomes.Type != "" {
			if err := appendSource("outcomes", dataset.Outcomes); err != nil {
				return "", err
			}
		}
		if dataset.Assignments.Type != "" {
			if err := appendSource("assignments", dataset.Assignments); err != nil {
				return "", err
			}
		}
	}

	sort.Strings(parts)
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:]), nil
}

func hashFile(path string) (string, error) {
	// #nosec G304 -- file path validated by caller
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
