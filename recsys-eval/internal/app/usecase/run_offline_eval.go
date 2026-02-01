package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics/distribution"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics/ranking"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/statistics"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/clock"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/reporting"
)

// OfflineEvalUsecase orchestrates offline evaluation runs.
type OfflineEvalUsecase struct {
	Exposures datasource.ExposureReader
	Outcomes  datasource.OutcomeReader
	Reporter  reporting.Writer
	Clock     clock.Clock
	Logger    logger.Logger
	Metadata  ReportMetadata
	Scale     ScaleConfig
}

func (u OfflineEvalUsecase) Run(ctx context.Context, evalCfg OfflineConfig, outputPath string, baselinePath string) (report.Report, error) {
	if u.Exposures == nil || u.Outcomes == nil {
		return report.Report{}, errors.New("exposure and outcome readers are required")
	}

	if strings.EqualFold(u.Scale.Mode, "stream") {
		return u.runStream(ctx, evalCfg, outputPath, baselinePath)
	}

	exposures, err := u.Exposures.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}
	outcomes, err := u.Outcomes.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}

	if evalCfg.TimeSplit != nil {
		_, testStart, testEnd, err := parseTimeSplit(*evalCfg.TimeSplit)
		if err != nil {
			return report.Report{}, err
		}
		exposures = dataset.FilterExposuresByTime(exposures, testStart, testEnd)
		outcomes = dataset.FilterOutcomesByTime(outcomes, testStart, testEnd)
	}

	joined, joinStats := dataset.JoinByRequest(exposures, outcomes)
	cases := make([]metrics.EvalCase, 0, len(joined))
	segCases := make(map[string][]metrics.EvalCase)

	for _, jc := range joined {
		recommended := make([]string, 0, len(jc.Exposure.Items))
		items := append([]dataset.ExposedItem(nil), jc.Exposure.Items...)
		sort.SliceStable(items, func(i, j int) bool { return items[i].Rank < items[j].Rank })
		for _, item := range items {
			recommended = append(recommended, item.ItemID)
		}

		relevant := map[string]struct{}{}
		for _, out := range jc.Outcomes {
			if strings.EqualFold(out.EventType, "click") || strings.EqualFold(out.EventType, "conversion") {
				relevant[out.ItemID] = struct{}{}
			}
		}

		c := metrics.EvalCase{Recommended: recommended, Relevant: relevant}
		cases = append(cases, c)

		segKey := dataset.SegmentKey(jc.Exposure.Context, evalCfg.SliceKeys)
		segCases[segKey] = append(segCases[segKey], c)
	}

	reg := metrics.NewRegistry()
	ranking.RegisterDefaults(reg)
	distribution.RegisterDefaults(reg)
	metricDefs, err := buildMetrics(reg, evalCfg.Metrics)
	if err != nil {
		return report.Report{}, err
	}

	bootstrapCfg := normalizeBootstrap(evalCfg.Bootstrap)
	metricsGlobal := aggregateMetrics(metricDefs, cases, bootstrapCfg)
	metricsBySegment := map[string][]report.MetricResult{}
	for seg, segCase := range segCases {
		metricsBySegment[seg] = aggregateMetrics(metricDefs, segCase, bootstrapCfg)
	}

	dq := buildDataQuality(exposures, outcomes, joinStats, 0)

	rep := report.Report{
		RunID:                   fmt.Sprintf("offline-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    "offline",
		CreatedAt:               u.Clock.Now().UTC(),
		Version:                 "0.1.0",
		BinaryVersion:           u.Metadata.BinaryVersion,
		GitCommit:               u.Metadata.GitCommit,
		EffectiveConfig:         u.Metadata.EffectiveConfig,
		InputDatasetFingerprint: u.Metadata.InputDatasetFingerprint,
		Artifacts:               u.Metadata.Artifacts,
		Summary: report.Summary{
			CasesEvaluated: len(cases),
		},
		Offline: &report.OfflineReport{
			Metrics:   metricsGlobal,
			BySegment: metricsBySegment,
		},
		DataQuality: &dq,
	}

	gateFailed := false
	if baselinePath != "" {
		baseline, err := readBaseline(baselinePath)
		if err != nil {
			return report.Report{}, err
		}
		rep.Offline.Baseline = compareBaseline(baseline, metricsGlobal)
		rep.Gates = evaluateGates(evalCfg.Gates, rep.Offline.Baseline)
		for _, g := range rep.Gates {
			if !g.Passed {
				u.Logger.Errorf("gate failed: %s delta=%.6f", g.Metric, g.Delta)
				gateFailed = true
			}
		}
	}

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, err
	}
	if evalCfg.HistoryDir != "" {
		if err := os.MkdirAll(evalCfg.HistoryDir, 0o750); err != nil {
			return report.Report{}, err
		}
		historyPath := filepath.Join(evalCfg.HistoryDir, rep.RunID+".json")
		if err := u.Reporter.Write(ctx, rep, historyPath); err != nil {
			return report.Report{}, err
		}
	}
	if gateFailed {
		return rep, fmt.Errorf("one or more regression gates failed")
	}

	return rep, nil
}

func (u OfflineEvalUsecase) runStream(ctx context.Context, evalCfg OfflineConfig, outputPath string, baselinePath string) (report.Report, error) {
	if evalCfg.TimeSplit != nil {
		return report.Report{}, fmt.Errorf("time_split is not supported in stream mode")
	}
	if evalCfg.Bootstrap != nil && evalCfg.Bootstrap.Enabled {
		return report.Report{}, fmt.Errorf("bootstrap is not supported in stream mode")
	}

	expStream, ok := u.Exposures.(datasource.ExposureStreamReader)
	if !ok {
		return report.Report{}, fmt.Errorf("exposure stream reader is required for stream mode")
	}
	outStream, ok := u.Outcomes.(datasource.OutcomeStreamReader)
	if !ok {
		return report.Report{}, fmt.Errorf("outcome stream reader is required for stream mode")
	}

	reg := metrics.NewRegistry()
	ranking.RegisterDefaults(reg)
	distribution.RegisterDefaults(reg)
	metricDefs, err := buildMetrics(reg, evalCfg.Metrics)
	if err != nil {
		return report.Report{}, err
	}
	for _, metric := range metricDefs {
		if _, ok := metric.(metrics.DatasetMetric); ok {
			return report.Report{}, fmt.Errorf("dataset metrics are not supported in stream mode")
		}
	}

	globalSums := make([]float64, len(metricDefs))
	globalCount := 0
	segSums := map[string][]float64{}
	segCounts := map[string]int{}

	joinRes, err := streamJoinByRequest(ctx, expStream, outStream, func(jc dataset.JoinedCase) error {
		recommended := make([]string, 0, len(jc.Exposure.Items))
		items := append([]dataset.ExposedItem(nil), jc.Exposure.Items...)
		sort.SliceStable(items, func(i, j int) bool { return items[i].Rank < items[j].Rank })
		for _, item := range items {
			recommended = append(recommended, item.ItemID)
		}

		relevant := map[string]struct{}{}
		for _, out := range jc.Outcomes {
			if strings.EqualFold(out.EventType, "click") || strings.EqualFold(out.EventType, "conversion") {
				relevant[out.ItemID] = struct{}{}
			}
		}

		c := metrics.EvalCase{Recommended: recommended, Relevant: relevant}
		for idx, metric := range metricDefs {
			globalSums[idx] += metric.Compute(c)
		}
		globalCount++

		segKey := dataset.SegmentKey(jc.Exposure.Context, evalCfg.SliceKeys)
		if _, ok := segSums[segKey]; !ok {
			segSums[segKey] = make([]float64, len(metricDefs))
		}
		for idx, metric := range metricDefs {
			segSums[segKey][idx] += metric.Compute(c)
		}
		segCounts[segKey]++
		return nil
	}, streamJoinOptions{MaxOpenRequests: u.Scale.Stream.MaxOpenRequests})
	if err != nil {
		return report.Report{}, err
	}

	metricsGlobal := make([]report.MetricResult, len(metricDefs))
	for i, metric := range metricDefs {
		mean := 0.0
		if globalCount > 0 {
			mean = globalSums[i] / float64(globalCount)
		}
		metricsGlobal[i] = report.MetricResult{Name: metric.Name(), Value: mean}
	}

	metricsBySegment := map[string][]report.MetricResult{}
	for seg, sums := range segSums {
		results := make([]report.MetricResult, len(metricDefs))
		count := segCounts[seg]
		for i, metric := range metricDefs {
			mean := 0.0
			if count > 0 {
				mean = sums[i] / float64(count)
			}
			results[i] = report.MetricResult{Name: metric.Name(), Value: mean}
		}
		metricsBySegment[seg] = results
	}

	dq := buildDataQualityFromCounts(joinRes.JoinStats.ExposureCount, joinRes.ExposureMissing, joinRes.JoinStats.OutcomeCount, joinRes.OutcomeMissing, joinRes.JoinStats, 0)

	rep := report.Report{
		RunID:                   fmt.Sprintf("offline-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    "offline",
		CreatedAt:               u.Clock.Now().UTC(),
		Version:                 "0.1.0",
		BinaryVersion:           u.Metadata.BinaryVersion,
		GitCommit:               u.Metadata.GitCommit,
		EffectiveConfig:         u.Metadata.EffectiveConfig,
		InputDatasetFingerprint: u.Metadata.InputDatasetFingerprint,
		Artifacts:               u.Metadata.Artifacts,
		Summary: report.Summary{
			CasesEvaluated: globalCount,
		},
		Offline: &report.OfflineReport{
			Metrics:   metricsGlobal,
			BySegment: metricsBySegment,
		},
		DataQuality: &dq,
	}

	gateFailed := false
	if baselinePath != "" {
		baseline, err := readBaseline(baselinePath)
		if err != nil {
			return report.Report{}, err
		}
		rep.Offline.Baseline = compareBaseline(baseline, metricsGlobal)
		rep.Gates = evaluateGates(evalCfg.Gates, rep.Offline.Baseline)
		for _, g := range rep.Gates {
			if !g.Passed {
				u.Logger.Errorf("gate failed: %s delta=%.6f", g.Metric, g.Delta)
				gateFailed = true
			}
		}
	}

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, err
	}
	if evalCfg.HistoryDir != "" {
		if err := os.MkdirAll(evalCfg.HistoryDir, 0o750); err != nil {
			return report.Report{}, err
		}
		historyPath := filepath.Join(evalCfg.HistoryDir, rep.RunID+".json")
		if err := u.Reporter.Write(ctx, rep, historyPath); err != nil {
			return report.Report{}, err
		}
	}
	if gateFailed {
		return rep, fmt.Errorf("one or more regression gates failed")
	}

	return rep, nil
}

func buildMetrics(reg *metrics.Registry, specs []metrics.MetricSpec) ([]metrics.Metric, error) {
	defs := make([]metrics.Metric, 0, len(specs))
	for _, spec := range specs {
		metric, err := reg.Build(spec)
		if err != nil {
			return nil, err
		}
		defs = append(defs, metric)
	}
	return defs, nil
}

func aggregateMetrics(metricDefs []metrics.Metric, cases []metrics.EvalCase, bootstrapCfg *BootstrapConfig) []report.MetricResult {
	results := make([]report.MetricResult, 0, len(metricDefs))
	for idx, metric := range metricDefs {
		if datasetMetric, ok := metric.(metrics.DatasetMetric); ok {
			value := datasetMetric.ComputeDataset(cases)
			results = append(results, report.MetricResult{Name: metric.Name(), Value: value})
			continue
		}
		values := make([]float64, len(cases))
		sum := 0.0
		for i, c := range cases {
			val := metric.Compute(c)
			values[i] = val
			sum += val
		}
		mean := 0.0
		if len(cases) > 0 {
			mean = sum / float64(len(cases))
		}
		result := report.MetricResult{Name: metric.Name(), Value: mean}
		if bootstrapCfg != nil && bootstrapCfg.Enabled && len(values) > 1 {
			seed := bootstrapCfg.Seed + int64(idx)
			ci := statistics.BootstrapMean(values, bootstrapCfg.Iterations, seed, bootstrapCfg.CILevel)
			if ci != nil {
				result.CI = &report.MetricCI{Lower: ci.Lower, Upper: ci.Upper, Level: ci.Level}
			}
		}
		results = append(results, result)
	}
	return results
}

func readBaseline(path string) (report.Report, error) {
	// #nosec G304 -- baseline path provided by operator
	data, err := os.ReadFile(path)
	if err != nil {
		return report.Report{}, err
	}
	var rep report.Report
	if err := json.Unmarshal(data, &rep); err != nil {
		return report.Report{}, err
	}
	if rep.Offline == nil {
		return report.Report{}, fmt.Errorf("baseline report missing offline section")
	}
	return rep, nil
}

func compareBaseline(baseline report.Report, current []report.MetricResult) *report.BaselineComparison {
	baseMetrics := map[string]report.MetricResult{}
	for _, m := range baseline.Offline.Metrics {
		baseMetrics[strings.ToLower(m.Name)] = m
	}
	var deltas []report.MetricDelta
	for _, m := range current {
		base, ok := baseMetrics[strings.ToLower(m.Name)]
		if !ok {
			continue
		}
		delta := m.Value - base.Value
		rel := 0.0
		if base.Value != 0 {
			rel = delta / base.Value
		}
		deltas = append(deltas, report.MetricDelta{Name: m.Name, Delta: delta, RelativeDelta: rel, Before: base.Value, After: m.Value})
	}
	return &report.BaselineComparison{BaselineRunID: baseline.RunID, Deltas: deltas}
}

func evaluateGates(gates []GateSpec, baseline *report.BaselineComparison) []report.GateResult {
	if baseline == nil {
		return nil
	}
	deltas := map[string]report.MetricDelta{}
	for _, d := range baseline.Deltas {
		deltas[strings.ToLower(d.Name)] = d
	}
	results := make([]report.GateResult, 0, len(gates))
	for _, g := range gates {
		d, ok := deltas[strings.ToLower(g.Metric)]
		if !ok {
			results = append(results, report.GateResult{Metric: g.Metric, MaxDrop: g.MaxDrop, Delta: 0, Passed: false})
			continue
		}
		passed := d.Delta >= -g.MaxDrop
		results = append(results, report.GateResult{Metric: g.Metric, MaxDrop: g.MaxDrop, Delta: d.Delta, Passed: passed})
	}
	return results
}

func buildDataQuality(exposures []dataset.Exposure, outcomes []dataset.Outcome, joinStats dataset.JoinStats, assignmentCount int) report.DataQualityReport {
	exposureMissing := 0
	for _, e := range exposures {
		if e.RequestID == "" || len(e.Items) == 0 {
			exposureMissing++
		}
	}
	outcomeMissing := 0
	for _, o := range outcomes {
		if o.RequestID == "" || o.ItemID == "" {
			outcomeMissing++
		}
	}
	return buildDataQualityFromCounts(len(exposures), exposureMissing, len(outcomes), outcomeMissing, joinStats, assignmentCount)
}

func buildDataQualityFromCounts(exposureCount, exposureMissing, outcomeCount, outcomeMissing int, joinStats dataset.JoinStats, assignmentCount int) report.DataQualityReport {
	exposureRate := missingRate(exposureMissing, exposureCount)
	outcomeRate := missingRate(outcomeMissing, outcomeCount)

	dq := report.DataQualityReport{
		ExposureCompleteness: report.CompletenessResult{Total: exposureCount, Missing: exposureMissing, MissingRate: exposureRate},
		OutcomeCompleteness:  report.CompletenessResult{Total: outcomeCount, Missing: outcomeMissing, MissingRate: outcomeRate},
		JoinIntegrity: report.JoinIntegrity{
			ExposureCount:    joinStats.ExposureCount,
			OutcomeCount:     joinStats.OutcomeCount,
			AssignmentsCount: assignmentCount,
		},
	}
	if joinStats.ExposureCount > 0 {
		dq.JoinIntegrity.ExposureJoinRate = float64(joinStats.ExposuresJoined) / float64(joinStats.ExposureCount)
	}
	if joinStats.OutcomeCount > 0 {
		dq.JoinIntegrity.OutcomeJoinRate = float64(joinStats.OutcomesJoined) / float64(joinStats.OutcomeCount)
	}
	if assignmentCount > 0 {
		dq.JoinIntegrity.AssignmentJoinRate = float64(joinStats.ExposuresJoined) / float64(assignmentCount)
	}
	return dq
}

func missingRate(missing, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(missing) / float64(total)
}

func normalizeBootstrap(cfg *BootstrapConfig) *BootstrapConfig {
	if cfg == nil {
		return nil
	}
	if cfg.Iterations <= 0 {
		cfg.Iterations = 200
	}
	if cfg.CILevel <= 0 || cfg.CILevel >= 1 {
		cfg.CILevel = 0.95
	}
	if cfg.Seed == 0 {
		cfg.Seed = 42
	}
	return cfg
}

func parseTimeSplit(split TimeSplitConfig) (trainStart, testStart, testEnd time.Time, err error) {
	trainStart, trainEnd, err := parseWindow(split.Train)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}
	testStart, testEnd, err = parseWindow(split.Test)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}
	if trainStart.IsZero() || trainEnd.IsZero() || testStart.IsZero() || testEnd.IsZero() {
		return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("time_split requires train.start, train.end, test.start, and test.end")
	}
	if !trainEnd.IsZero() && !testStart.IsZero() && trainEnd.After(testStart) {
		return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("train window must end before test window starts")
	}
	return trainStart, testStart, testEnd, nil
}

func parseWindow(win TimeWindow) (time.Time, time.Time, error) {
	start, err := parseTime(win.Start)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseTime(win.End)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("window start must be before end")
	}
	return start, end, nil
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if ts, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return ts, nil
	}
	return time.Parse(time.RFC3339, value)
}
