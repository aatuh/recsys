package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/clock"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/reporting"
)

// AACheckUsecase runs A/A sanity checks on experiment data.
type AACheckUsecase struct {
	Exposures   datasource.ExposureReader
	Outcomes    datasource.OutcomeReader
	Assignments datasource.AssignmentReader
	Reporter    reporting.Writer
	Clock       clock.Clock
	Logger      logger.Logger
	Metadata    ReportMetadata
}

func (u AACheckUsecase) Run(ctx context.Context, cfg ExperimentConfig, outputPath string) (report.Report, error) {
	if u.Exposures == nil || u.Outcomes == nil || u.Assignments == nil {
		return report.Report{}, errors.New("exposure, outcome, and assignment readers are required")
	}
	if cfg.AACheck == nil || !cfg.AACheck.Enabled {
		return report.Report{}, errors.New("aa_check config must be enabled")
	}

	exposures, err := u.Exposures.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}
	outcomes, err := u.Outcomes.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}
	assignments, err := u.Assignments.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}

	joined, joinStats := dataset.JoinByRequest(exposures, outcomes)

	assignByReq := map[string]dataset.Assignment{}
	assignmentMissing := 0
	for _, a := range assignments {
		if a.RequestID == "" || a.Variant == "" {
			assignmentMissing++
			continue
		}
		if cfg.ExperimentID != "" && !strings.EqualFold(a.ExperimentID, cfg.ExperimentID) {
			continue
		}
		assignByReq[a.RequestID] = a
	}

	variantAgg := map[string]*variantAccumulator{}
	segAgg := map[string]map[string]*variantAccumulator{}

	assignedExposures := 0
	for reqID, jc := range joined {
		assignment, ok := assignByReq[reqID]
		if !ok {
			continue
		}
		assignedExposures++
		variant := assignment.Variant
		acc := getVariant(variantAgg, variant)

		outMetrics := computeOutcomeMetrics(jc.Outcomes)
		acc.addExposure(jc.Exposure, outMetrics)

		segKey := dataset.SegmentKey(jc.Exposure.Context, cfg.SliceKeys)
		if _, ok := segAgg[segKey]; !ok {
			segAgg[segKey] = map[string]*variantAccumulator{}
		}
		segAcc := getVariant(segAgg[segKey], variant)
		segAcc.addExposure(jc.Exposure, outMetrics)
	}

	significance := buildSignificance(cfg, variantAgg, segAgg)
	thresholds := cfg.AACheck.Thresholds
	summaries, totalTests, uniformish := summarizePValues(significance, thresholds)

	dq := buildDataQuality(exposures, outcomes, joinStats, len(assignByReq))
	dq.AssignmentCompleteness = &report.CompletenessResult{
		Total:       len(assignments),
		Missing:     assignmentMissing,
		MissingRate: missingRate(assignmentMissing, len(assignments)),
	}
	dq.JoinIntegrity.AssignmentJoinRate = 0
	if len(assignByReq) > 0 {
		dq.JoinIntegrity.AssignmentJoinRate = float64(assignedExposures) / float64(len(assignByReq))
	}

	warnings := []string{}
	if !uniformish {
		warnings = append(warnings, "aa_check_false_positives")
	}

	rep := report.Report{
		RunID:                   fmt.Sprintf("aa-check-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    report.ModeAACheck,
		CreatedAt:               u.Clock.Now().UTC(),
		Version:                 "0.1.0",
		BinaryVersion:           u.Metadata.BinaryVersion,
		GitCommit:               u.Metadata.GitCommit,
		EffectiveConfig:         u.Metadata.EffectiveConfig,
		InputDatasetFingerprint: u.Metadata.InputDatasetFingerprint,
		Artifacts:               u.Metadata.Artifacts,
		Summary: report.Summary{
			CasesEvaluated: assignedExposures,
		},
		AA: &report.AAReport{
			Metrics:    summaries,
			Thresholds: thresholds,
			TotalTests: totalTests,
			Uniformish: uniformish,
		},
		DataQuality: &dq,
		Warnings:    warnings,
	}
	rep.Summary.Executive = buildExecutiveSummary(rep)

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, err
	}

	return rep, nil
}

func summarizePValues(tests []report.ExperimentSignificance, thresholds []float64) ([]report.AAPValueSummary, int, bool) {
	if len(thresholds) == 0 {
		thresholds = []float64{0.05, 0.01, 0.001}
	}
	global := map[string][]float64{}
	segmented := map[string][]float64{}
	all := make([]float64, 0, len(tests))

	for _, t := range tests {
		all = append(all, t.PValue)
		if t.Segment == "" {
			global[t.Metric] = append(global[t.Metric], t.PValue)
		} else {
			segmented[t.Metric] = append(segmented[t.Metric], t.PValue)
		}
	}

	summaries := make([]report.AAPValueSummary, 0, len(global)+len(segmented))
	appendSummaries := func(bySegment bool, source map[string][]float64) {
		metrics := make([]string, 0, len(source))
		for metric := range source {
			metrics = append(metrics, metric)
		}
		sort.Strings(metrics)
		for _, metric := range metrics {
			pvals := source[metric]
			summaries = append(summaries, report.AAPValueSummary{
				Metric:    metric,
				BySegment: bySegment,
				Counts:    countPValues(pvals, thresholds),
			})
		}
	}

	appendSummaries(false, global)
	if len(segmented) > 0 {
		appendSummaries(true, segmented)
	}

	total := len(all)
	uniformish := looksUniform(all, thresholds)
	return summaries, total, uniformish
}

func countPValues(pvals []float64, thresholds []float64) map[string]int {
	counts := map[string]int{}
	for _, t := range thresholds {
		counts[formatThreshold(t)] = 0
	}
	for _, p := range pvals {
		for _, t := range thresholds {
			if p < t {
				key := formatThreshold(t)
				counts[key]++
			}
		}
	}
	return counts
}

func looksUniform(pvals []float64, thresholds []float64) bool {
	total := len(pvals)
	if total == 0 {
		return true
	}
	counts := countPValues(pvals, thresholds)
	for _, t := range thresholds {
		expected := float64(total) * t
		if expected < 1 {
			continue
		}
		actual := float64(counts[formatThreshold(t)])
		if actual > expected*2.5 {
			return false
		}
	}
	return true
}

func formatThreshold(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", v), "0"), ".")
}
