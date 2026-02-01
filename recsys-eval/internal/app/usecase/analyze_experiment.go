package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/decision"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/statistics"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/clock"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/reporting"
)

// ExperimentUsecase orchestrates experiment analysis.
type ExperimentUsecase struct {
	Exposures   datasource.ExposureReader
	Outcomes    datasource.OutcomeReader
	Assignments datasource.AssignmentReader
	Reporter    reporting.Writer
	Decision    reporting.DecisionWriter
	Clock       clock.Clock
	Logger      logger.Logger
	Metadata    ReportMetadata
	Scale       ScaleConfig
}

var errDecisionDisabled = errors.New("decision disabled")

func (u ExperimentUsecase) Run(ctx context.Context, evalCfg ExperimentConfig, outputPath string) (report.Report, error) {
	rep, _, err := u.run(ctx, evalCfg, outputPath)
	return rep, err
}

func (u ExperimentUsecase) RunWithDecision(ctx context.Context, evalCfg ExperimentConfig, outputPath string) (*decision.Artifact, error) {
	_, artifact, err := u.run(ctx, evalCfg, outputPath)
	return artifact, err
}

func (u ExperimentUsecase) run(ctx context.Context, evalCfg ExperimentConfig, outputPath string) (report.Report, *decision.Artifact, error) {
	if u.Exposures == nil || u.Outcomes == nil || u.Assignments == nil {
		return report.Report{}, nil, errors.New("exposure, outcome, and assignment readers are required")
	}

	if strings.EqualFold(u.Scale.Mode, "stream") {
		return u.runStream(ctx, evalCfg, outputPath)
	}

	exposures, err := u.Exposures.Read(ctx)
	if err != nil {
		return report.Report{}, nil, err
	}
	outcomes, err := u.Outcomes.Read(ctx)
	if err != nil {
		return report.Report{}, nil, err
	}
	assignments, err := u.Assignments.Read(ctx)
	if err != nil {
		return report.Report{}, nil, err
	}

	joined, joinStats := dataset.JoinByRequest(exposures, outcomes)

	assignByReq := map[string]dataset.Assignment{}
	assignmentMissing := 0
	for _, a := range assignments {
		if a.RequestID == "" || a.Variant == "" {
			assignmentMissing++
			continue
		}
		if evalCfg.ExperimentID != "" && !strings.EqualFold(a.ExperimentID, evalCfg.ExperimentID) {
			continue
		}
		assignByReq[a.RequestID] = a
	}

	variantAgg := map[string]*variantAccumulator{}
	segAgg := map[string]map[string]*variantAccumulator{}

	cupedCfg := normalizeCuped(evalCfg.CUPED)
	globalCuped := cupedGlobal{}
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
		clickIndicator := 0.0
		if outMetrics.clicks > 0 {
			clickIndicator = 1
		}
		convIndicator := 0.0
		if outMetrics.conversions > 0 {
			convIndicator = 1
		}

		acc.addExposure(jc.Exposure, outMetrics)

		if cupedCfg.Enabled && cupedCfg.CovariateKey != "" {
			if cov, ok := parseCovariate(jc.Exposure.Context, cupedCfg.CovariateKey); ok {
				acc.addCovariate(cov, clickIndicator, convIndicator)
				globalCuped.add(cov, clickIndicator, convIndicator)
			}
		}

		segKey := dataset.SegmentKey(jc.Exposure.Context, evalCfg.SliceKeys)
		if _, ok := segAgg[segKey]; !ok {
			segAgg[segKey] = map[string]*variantAccumulator{}
		}
		segAcc := getVariant(segAgg[segKey], variant)
		segAcc.addExposure(jc.Exposure, outMetrics)
		if cupedCfg.Enabled && cupedCfg.CovariateKey != "" {
			if cov, ok := parseCovariate(jc.Exposure.Context, cupedCfg.CovariateKey); ok {
				segAcc.addCovariate(cov, clickIndicator, convIndicator)
			}
		}
	}

	cupedParams := buildCupedParams(cupedCfg, globalCuped, assignedExposures)
	variants := buildVariantMetrics(variantAgg, cupedParams)
	segVariants := map[string][]report.VariantMetrics{}
	for seg, vmap := range segAgg {
		segVariants[seg] = buildVariantMetrics(vmap, cupedParams)
	}

	guardrails := buildGuardrailReport(variantAgg, evalCfg.Guardrails)
	significance := buildSignificance(evalCfg, variantAgg, segAgg)
	power := buildPowerReport(evalCfg.Power)
	srm := buildSRMReport(evalCfg.SRM, assignments, evalCfg.ExperimentID)

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

	rep := report.Report{
		RunID:                   fmt.Sprintf("experiment-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    report.ModeExperiment,
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
		Experiment: &report.ExperimentReport{
			ExperimentID: evalCfg.ExperimentID,
			Variants:     variants,
			BySegment:    segVariants,
			Guardrails:   guardrails,
			Cuped:        cupedParams.report,
			Power:        power,
			Significance: significance,
			SRM:          srm,
		},
		DataQuality: &dq,
	}

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, nil, err
	}

	artifact, err := u.maybeWriteDecision(ctx, evalCfg, rep, outputPath, guardrails, srm, exposures, outcomes, nil)
	if err != nil {
		if errors.Is(err, errDecisionDisabled) {
			return rep, nil, nil
		}
		return report.Report{}, nil, err
	}

	return rep, artifact, nil
}

func (u ExperimentUsecase) runStream(ctx context.Context, evalCfg ExperimentConfig, outputPath string) (report.Report, *decision.Artifact, error) {
	expStream, ok := u.Exposures.(datasource.ExposureStreamReader)
	if !ok {
		return report.Report{}, nil, fmt.Errorf("exposure stream reader is required for stream mode")
	}
	outStream, ok := u.Outcomes.(datasource.OutcomeStreamReader)
	if !ok {
		return report.Report{}, nil, fmt.Errorf("outcome stream reader is required for stream mode")
	}

	assignments, err := u.Assignments.Read(ctx)
	if err != nil {
		return report.Report{}, nil, err
	}

	assignByReq := map[string]dataset.Assignment{}
	assignmentMissing := 0
	for _, a := range assignments {
		if a.RequestID == "" || a.Variant == "" {
			assignmentMissing++
			continue
		}
		if evalCfg.ExperimentID != "" && !strings.EqualFold(a.ExperimentID, evalCfg.ExperimentID) {
			continue
		}
		assignByReq[a.RequestID] = a
	}

	variantAgg := map[string]*variantAccumulator{}
	segAgg := map[string]map[string]*variantAccumulator{}

	cupedCfg := normalizeCuped(evalCfg.CUPED)
	globalCuped := cupedGlobal{}
	assignedExposures := 0

	joinRes, err := streamJoinByRequest(ctx, expStream, outStream, func(jc dataset.JoinedCase) error {
		assignment, ok := assignByReq[jc.Exposure.RequestID]
		if !ok {
			return nil
		}
		assignedExposures++
		variant := assignment.Variant
		acc := getVariant(variantAgg, variant)

		outMetrics := computeOutcomeMetrics(jc.Outcomes)
		clickIndicator := 0.0
		if outMetrics.clicks > 0 {
			clickIndicator = 1
		}
		convIndicator := 0.0
		if outMetrics.conversions > 0 {
			convIndicator = 1
		}

		acc.addExposure(jc.Exposure, outMetrics)

		if cupedCfg.Enabled && cupedCfg.CovariateKey != "" {
			if cov, ok := parseCovariate(jc.Exposure.Context, cupedCfg.CovariateKey); ok {
				acc.addCovariate(cov, clickIndicator, convIndicator)
				globalCuped.add(cov, clickIndicator, convIndicator)
			}
		}

		segKey := dataset.SegmentKey(jc.Exposure.Context, evalCfg.SliceKeys)
		if _, ok := segAgg[segKey]; !ok {
			segAgg[segKey] = map[string]*variantAccumulator{}
		}
		segAcc := getVariant(segAgg[segKey], variant)
		segAcc.addExposure(jc.Exposure, outMetrics)
		if cupedCfg.Enabled && cupedCfg.CovariateKey != "" {
			if cov, ok := parseCovariate(jc.Exposure.Context, cupedCfg.CovariateKey); ok {
				segAcc.addCovariate(cov, clickIndicator, convIndicator)
			}
		}
		return nil
	}, streamJoinOptions{MaxOpenRequests: u.Scale.Stream.MaxOpenRequests})
	if err != nil {
		return report.Report{}, nil, err
	}

	cupedParams := buildCupedParams(cupedCfg, globalCuped, assignedExposures)
	variants := buildVariantMetrics(variantAgg, cupedParams)
	segVariants := map[string][]report.VariantMetrics{}
	for seg, vmap := range segAgg {
		segVariants[seg] = buildVariantMetrics(vmap, cupedParams)
	}

	guardrails := buildGuardrailReport(variantAgg, evalCfg.Guardrails)
	significance := buildSignificance(evalCfg, variantAgg, segAgg)
	power := buildPowerReport(evalCfg.Power)
	srm := buildSRMReport(evalCfg.SRM, assignments, evalCfg.ExperimentID)

	dq := buildDataQualityFromCounts(joinRes.JoinStats.ExposureCount, joinRes.ExposureMissing, joinRes.JoinStats.OutcomeCount, joinRes.OutcomeMissing, joinRes.JoinStats, len(assignByReq))
	dq.AssignmentCompleteness = &report.CompletenessResult{
		Total:       len(assignments),
		Missing:     assignmentMissing,
		MissingRate: missingRate(assignmentMissing, len(assignments)),
	}
	dq.JoinIntegrity.AssignmentJoinRate = 0
	if len(assignByReq) > 0 {
		dq.JoinIntegrity.AssignmentJoinRate = float64(assignedExposures) / float64(len(assignByReq))
	}

	rep := report.Report{
		RunID:                   fmt.Sprintf("experiment-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    report.ModeExperiment,
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
		Experiment: &report.ExperimentReport{
			ExperimentID: evalCfg.ExperimentID,
			Variants:     variants,
			BySegment:    segVariants,
			Guardrails:   guardrails,
			Cuped:        cupedParams.report,
			Power:        power,
			Significance: significance,
			SRM:          srm,
		},
		DataQuality: &dq,
	}

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, nil, err
	}

	window := datasetWindowFromStream(joinRes)
	artifact, err := u.maybeWriteDecision(ctx, evalCfg, rep, outputPath, guardrails, srm, nil, nil, &window)
	if err != nil {
		if errors.Is(err, errDecisionDisabled) {
			return rep, nil, nil
		}
		return report.Report{}, nil, err
	}

	return rep, artifact, nil
}

type outcomeMetrics struct {
	clicks      int
	conversions int
	revenue     float64
}

func computeOutcomeMetrics(outcomes []dataset.Outcome) outcomeMetrics {
	m := outcomeMetrics{}
	for _, o := range outcomes {
		switch strings.ToLower(o.EventType) {
		case "click":
			m.clicks++
		case "conversion":
			m.conversions++
			m.revenue += o.Value
		}
	}
	return m
}

type variantAccumulator struct {
	variant                string
	exposures              int
	clicks                 int
	conversions            int
	revenue                float64
	latencies              []float64
	errors                 int
	empty                  int
	minTS                  time.Time
	maxTS                  time.Time
	covariateSum           float64
	covariateCount         int
	clickIndicatorSum      float64
	conversionIndicatorSum float64
}

func getVariant(m map[string]*variantAccumulator, variant string) *variantAccumulator {
	acc, ok := m[variant]
	if !ok {
		acc = &variantAccumulator{variant: variant}
		m[variant] = acc
	}
	return acc
}

func (v *variantAccumulator) addExposure(exp dataset.Exposure, out outcomeMetrics) {
	v.exposures++
	v.clicks += out.clicks
	v.conversions += out.conversions
	v.revenue += out.revenue
	if !exp.Timestamp.IsZero() {
		if v.minTS.IsZero() || exp.Timestamp.Before(v.minTS) {
			v.minTS = exp.Timestamp
		}
		if v.maxTS.IsZero() || exp.Timestamp.After(v.maxTS) {
			v.maxTS = exp.Timestamp
		}
	}
	if exp.LatencyMs != nil {
		v.latencies = append(v.latencies, *exp.LatencyMs)
	}
	if exp.Error != nil && *exp.Error {
		v.errors++
	}
	if len(exp.Items) == 0 {
		v.empty++
	}
}

func (v *variantAccumulator) addCovariate(x, clickIndicator, convIndicator float64) {
	v.covariateSum += x
	v.covariateCount++
	v.clickIndicatorSum += clickIndicator
	v.conversionIndicatorSum += convIndicator
}

type cupedConfig struct {
	Enabled      bool
	CovariateKey string
}

type cupedGlobal struct {
	count      int
	sumX       float64
	sumX2      float64
	sumYClick  float64
	sumYConv   float64
	sumXYClick float64
	sumXYConv  float64
}

func (c *cupedGlobal) add(x, yClick, yConv float64) {
	c.count++
	c.sumX += x
	c.sumX2 += x * x
	c.sumYClick += yClick
	c.sumYConv += yConv
	c.sumXYClick += x * yClick
	c.sumXYConv += x * yConv
}

type cupedParams struct {
	enabled  bool
	meanX    float64
	thetaCTR float64
	thetaCVR float64
	report   *report.CupedReport
}

func normalizeCuped(cfg *CUPEDConfig) cupedConfig {
	if cfg == nil {
		return cupedConfig{}
	}
	return cupedConfig{Enabled: cfg.Enabled, CovariateKey: cfg.CovariateKey}
}

func buildCupedParams(cfg cupedConfig, global cupedGlobal, total int) *cupedParams {
	if !cfg.Enabled || cfg.CovariateKey == "" || global.count == 0 {
		return &cupedParams{enabled: false, report: nil}
	}
	meanX := global.sumX / float64(global.count)
	meanYClick := global.sumYClick / float64(global.count)
	meanYConv := global.sumYConv / float64(global.count)
	varX := global.sumX2/float64(global.count) - meanX*meanX
	covClick := global.sumXYClick/float64(global.count) - meanX*meanYClick
	covConv := global.sumXYConv/float64(global.count) - meanX*meanYConv
	if varX <= 0 {
		return &cupedParams{enabled: false, report: nil}
	}
	thetaCTR := covClick / varX
	thetaCVR := covConv / varX
	coverage := 0.0
	if total > 0 {
		coverage = float64(global.count) / float64(total)
	}

	return &cupedParams{
		enabled:  true,
		meanX:    meanX,
		thetaCTR: thetaCTR,
		thetaCVR: thetaCVR,
		report: &report.CupedReport{
			Enabled:      true,
			CovariateKey: cfg.CovariateKey,
			ThetaCTR:     thetaCTR,
			ThetaCVR:     thetaCVR,
			CoverageRate: coverage,
		},
	}
}

func parseCovariate(ctx map[string]string, key string) (float64, bool) {
	if ctx == nil {
		return 0, false
	}
	value, ok := ctx[key]
	if !ok || value == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func buildVariantMetrics(vmap map[string]*variantAccumulator, cuped *cupedParams) []report.VariantMetrics {
	variants := make([]report.VariantMetrics, 0, len(vmap))
	for _, acc := range vmap {
		ctr := rate(acc.clicks, acc.exposures)
		cvr := rate(acc.conversions, acc.exposures)
		revenuePerReq := 0.0
		if acc.exposures > 0 {
			revenuePerReq = acc.revenue / float64(acc.exposures)
		}
		throughput := 0.0
		if !acc.minTS.IsZero() && !acc.maxTS.IsZero() && acc.maxTS.After(acc.minTS) {
			seconds := acc.maxTS.Sub(acc.minTS).Seconds()
			if seconds > 0 {
				throughput = float64(acc.exposures) / seconds
			}
		}
		metric := report.VariantMetrics{
			Variant:           acc.variant,
			Exposures:         acc.exposures,
			Clicks:            acc.clicks,
			Conversions:       acc.conversions,
			CTR:               ctr,
			ConversionRate:    cvr,
			Revenue:           acc.revenue,
			RevenuePerRequest: revenuePerReq,
			ThroughputRPS:     throughput,
		}

		if cuped != nil && cuped.enabled && acc.covariateCount > 0 {
			meanXGroup := acc.covariateSum / float64(acc.covariateCount)
			meanClickGroup := acc.clickIndicatorSum / float64(acc.covariateCount)
			meanConvGroup := acc.conversionIndicatorSum / float64(acc.covariateCount)
			adjCTR := meanClickGroup - cuped.thetaCTR*(meanXGroup-cuped.meanX)
			adjCVR := meanConvGroup - cuped.thetaCVR*(meanXGroup-cuped.meanX)
			metric.CTRAdjusted = &adjCTR
			metric.ConversionRateAdjusted = &adjCVR
		}

		variants = append(variants, metric)
	}
	sort.Slice(variants, func(i, j int) bool { return variants[i].Variant < variants[j].Variant })
	return variants
}

func buildGuardrailReport(vmap map[string]*variantAccumulator, cfg GuardrailConfig) report.GuardrailReport {
	perVariant := make([]report.GuardrailMetrics, 0, len(vmap))
	for _, acc := range vmap {
		latencyP95 := percentile(acc.latencies, 0.95)
		errorRate := rate(acc.errors, acc.exposures)
		emptyRate := rate(acc.empty, acc.exposures)
		perVariant = append(perVariant, report.GuardrailMetrics{
			Variant:       acc.variant,
			LatencyP95Ms:  latencyP95,
			ErrorRate:     errorRate,
			EmptyRate:     emptyRate,
			PassLatency:   cfg.MaxLatencyP95Ms == 0 || latencyP95 <= cfg.MaxLatencyP95Ms,
			PassErrorRate: cfg.MaxErrorRate == 0 || errorRate <= cfg.MaxErrorRate,
			PassEmptyRate: cfg.MaxEmptyRate == 0 || emptyRate <= cfg.MaxEmptyRate,
		})
	}
	sort.Slice(perVariant, func(i, j int) bool { return perVariant[i].Variant < perVariant[j].Variant })
	return report.GuardrailReport{PerVariant: perVariant}
}

func buildSignificance(cfg ExperimentConfig, overall map[string]*variantAccumulator, segments map[string]map[string]*variantAccumulator) []report.ExperimentSignificance {
	control := selectControlVariant(cfg.ControlVariant, overall)
	if control == "" {
		return nil
	}
	primary := map[string]struct{}{}
	for _, m := range cfg.PrimaryMetrics {
		primary[strings.ToLower(m)] = struct{}{}
	}

	var tests []report.ExperimentSignificance
	addTests := func(segment string, vmap map[string]*variantAccumulator) {
		controlAcc, ok := vmap[control]
		if !ok {
			return
		}
		variants := make([]string, 0, len(vmap))
		for variant := range vmap {
			if variant == control {
				continue
			}
			variants = append(variants, variant)
		}
		sort.Strings(variants)
		for _, variant := range variants {
			acc := vmap[variant]
			pCTR := statistics.TwoProportionZTestPValue(acc.clicks, acc.exposures, controlAcc.clicks, controlAcc.exposures)
			tests = append(tests, report.ExperimentSignificance{
				Metric:  "ctr",
				Variant: variant,
				Segment: segment,
				PValue:  pCTR,
				Primary: isPrimary(primary, "ctr"),
			})
			pCVR := statistics.TwoProportionZTestPValue(acc.conversions, acc.exposures, controlAcc.conversions, controlAcc.exposures)
			tests = append(tests, report.ExperimentSignificance{
				Metric:  "conversion_rate",
				Variant: variant,
				Segment: segment,
				PValue:  pCVR,
				Primary: isPrimary(primary, "conversion_rate"),
			})
		}
	}

	addTests("", overall)
	segKeys := make([]string, 0, len(segments))
	for seg := range segments {
		segKeys = append(segKeys, seg)
	}
	sort.Strings(segKeys)
	for _, seg := range segKeys {
		addTests(seg, segments[seg])
	}

	cfgMC := normalizeMultipleComparisons(cfg.MultipleComparisons)
	return applyCorrection(tests, cfgMC)
}

func normalizeMultipleComparisons(cfg *MultipleComparisonsConfig) MultipleComparisonsConfig {
	if cfg == nil {
		return MultipleComparisonsConfig{Method: "fdr_bh", Alpha: 0.05}
	}
	method := cfg.Method
	if method == "" {
		method = "fdr_bh"
	}
	alpha := cfg.Alpha
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.05
	}
	return MultipleComparisonsConfig{Method: strings.ToLower(method), Alpha: alpha}
}

func applyCorrection(tests []report.ExperimentSignificance, cfg MultipleComparisonsConfig) []report.ExperimentSignificance {
	if len(tests) == 0 {
		return nil
	}

	method := cfg.Method
	alpha := cfg.Alpha

	switch method {
	case "none":
		for i := range tests {
			tests[i].AdjustedPValue = tests[i].PValue
			tests[i].Significant = tests[i].AdjustedPValue <= alpha
			tests[i].Correction = method
		}
		return tests
	case "bonferroni":
		m := float64(len(tests))
		for i := range tests {
			adj := tests[i].PValue * m
			if adj > 1 {
				adj = 1
			}
			tests[i].AdjustedPValue = adj
			tests[i].Significant = adj <= alpha
			tests[i].Correction = method
		}
		return tests
	default:
		// Benjamini-Hochberg FDR control.
		type entry struct {
			idx int
			p   float64
		}
		entries := make([]entry, len(tests))
		for i, t := range tests {
			entries[i] = entry{idx: i, p: t.PValue}
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].p < entries[j].p })
		m := float64(len(entries))
		adj := make([]float64, len(entries))
		for i := len(entries) - 1; i >= 0; i-- {
			rank := float64(i + 1)
			val := entries[i].p * m / rank
			if i == len(entries)-1 {
				adj[i] = val
			} else {
				adj[i] = math.Min(val, adj[i+1])
			}
		}
		for i, e := range entries {
			adjusted := adj[i]
			if adjusted > 1 {
				adjusted = 1
			}
			tests[e.idx].AdjustedPValue = adjusted
			tests[e.idx].Significant = adjusted <= alpha
			tests[e.idx].Correction = "fdr_bh"
		}
		return tests
	}
}

func selectControlVariant(control string, vmap map[string]*variantAccumulator) string {
	if control != "" {
		if _, ok := vmap[control]; ok {
			return control
		}
	}
	var chosen string
	max := -1
	for name, acc := range vmap {
		if acc.exposures > max {
			max = acc.exposures
			chosen = name
		}
	}
	return chosen
}

func isPrimary(primary map[string]struct{}, metric string) bool {
	_, ok := primary[strings.ToLower(metric)]
	return ok
}

func buildPowerReport(cfg *PowerConfig) *report.PowerAnalysisReport {
	if cfg == nil {
		return nil
	}
	baseline := cfg.BaselineRate
	if baseline <= 0 || baseline >= 1 || cfg.MDE <= 0 {
		return nil
	}
	metric := cfg.Metric
	if metric == "" {
		metric = "ctr"
	}
	alpha := cfg.Alpha
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.05
	}
	power := cfg.Power
	if power <= 0 || power >= 1 {
		power = 0.8
	}
	p0 := baseline
	p1 := p0 + cfg.MDE
	if p1 <= 0 || p1 >= 1 {
		return nil
	}
	zAlpha := statistics.InvNorm(1 - alpha/2)
	zBeta := statistics.InvNorm(power)
	pbar := (p0 + p1) / 2
	numerator := zAlpha*math.Sqrt(2*pbar*(1-pbar)) + zBeta*math.Sqrt(p0*(1-p0)+p1*(1-p1))
	n := (numerator * numerator) / ((p1 - p0) * (p1 - p0))
	if n <= 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return nil
	}
	perVariant := int(math.Ceil(n))
	estDays := 0.0
	if cfg.DailySamples > 0 {
		estDays = float64(perVariant*2) / float64(cfg.DailySamples)
	}

	return &report.PowerAnalysisReport{
		Metric:               strings.ToLower(metric),
		BaselineRate:         baseline,
		MDE:                  cfg.MDE,
		Alpha:                alpha,
		Power:                power,
		SampleSizePerVariant: perVariant,
		EstimatedDays:        estDays,
	}
}

func buildSRMReport(cfg *SRMConfig, assignments []dataset.Assignment, experimentID string) *report.SRMReport {
	if cfg == nil || !cfg.Enabled {
		return nil
	}
	alpha := cfg.Alpha
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.001
	}
	minSample := cfg.MinSample
	if minSample < 0 {
		minSample = 0
	}

	filtered := make([]dataset.Assignment, 0, len(assignments))
	for _, a := range assignments {
		if a.RequestID == "" || a.Variant == "" {
			continue
		}
		if experimentID != "" && !strings.EqualFold(a.ExperimentID, experimentID) {
			continue
		}
		filtered = append(filtered, a)
	}

	global := computeSRMResult(filtered, cfg.Expected, alpha, minSample, nil)
	bySegment := map[string]report.SRMResult{}
	if len(cfg.SliceKeys) > 0 {
		segMap := map[string][]dataset.Assignment{}
		for _, a := range filtered {
			segKey := dataset.SegmentKey(a.Context, cfg.SliceKeys)
			segMap[segKey] = append(segMap[segKey], a)
		}
		for seg, items := range segMap {
			bySegment[seg] = computeSRMResult(items, cfg.Expected, alpha, minSample, cfg.SliceKeys)
		}
	}

	rep := &report.SRMReport{
		Alpha:     alpha,
		MinSample: minSample,
		Global:    global,
	}
	if len(bySegment) > 0 {
		rep.BySegment = bySegment
	}
	return rep
}

func computeSRMResult(assignments []dataset.Assignment, expected map[string]float64, alpha float64, minSample int, _ []string) report.SRMResult {
	observed := map[string]int{}
	total := 0
	for _, a := range assignments {
		observed[a.Variant]++
		total++
	}
	if len(expected) == 0 {
		expected = map[string]float64{}
		for variant := range observed {
			expected[variant] = 1
		}
	}

	pValue, err := statistics.ChiSquarePValue(observed, expected)
	if err != nil {
		pValue = 1
	}
	detected := total >= minSample && pValue < alpha
	return report.SRMResult{
		PValue:     pValue,
		Observed:   observed,
		Expected:   expected,
		Detected:   detected,
		SampleSize: total,
	}
}

func (u ExperimentUsecase) maybeWriteDecision(ctx context.Context, cfg ExperimentConfig, rep report.Report, outputPath string, guardrails report.GuardrailReport, srm *report.SRMReport, exposures []dataset.Exposure, outcomes []dataset.Outcome, window *decision.Window) (*decision.Artifact, error) {
	if cfg.Decision == nil || !cfg.Decision.Enabled {
		return nil, errDecisionDisabled
	}
	if u.Decision == nil {
		return nil, errors.New("decision writer is required when decision is enabled")
	}
	artifact := buildDecisionArtifact(cfg, rep, guardrails, srm, exposures, outcomes, window)
	decisionPath := cfg.Decision.OutputPath
	if decisionPath == "" {
		decisionPath = outputPath + ".decision.json"
	}
	if err := u.Decision.Write(ctx, *artifact, decisionPath); err != nil {
		return nil, err
	}
	return artifact, nil
}

func buildDecisionArtifact(cfg ExperimentConfig, rep report.Report, guardrails report.GuardrailReport, srm *report.SRMReport, exposures []dataset.Exposure, outcomes []dataset.Outcome, window *decision.Window) *decision.Artifact {
	control := selectControlVariant(cfg.ControlVariant, toAccumulatorMap(rep.Experiment.Variants))
	candidate := selectCandidateVariant(cfg, rep.Experiment.Variants, control)

	primary := cfg.PrimaryMetrics
	if len(primary) == 0 {
		primary = []string{"ctr"}
	}

	thresholds := cfg.Decision.PrimaryThresholds
	if thresholds == nil {
		thresholds = map[string]float64{}
	}

	controlMetrics := findVariant(rep.Experiment.Variants, control)
	candidateMetrics := findVariant(rep.Experiment.Variants, candidate)

	metricDecisions := make([]decision.MetricDecision, 0, len(primary))
	primaryFail := false
	for _, metric := range primary {
		metricName := strings.ToLower(metric)
		before := metricValue(controlMetrics, metricName)
		after := metricValue(candidateMetrics, metricName)
		delta := after - before
		rel := 0.0
		if before != 0 {
			rel = delta / before
		}
		threshold := thresholds[metricName]
		if delta < threshold {
			primaryFail = true
		}
		metricDecisions = append(metricDecisions, decision.MetricDecision{
			Metric:        metricName,
			Delta:         delta,
			RelativeDelta: rel,
			Threshold:     threshold,
		})
	}

	guardrailDecisions := buildGuardrailDecisions(cfg.Guardrails, guardrails, candidate)
	guardrailFail := cfg.Decision.GuardrailGate && anyGuardrailFail(guardrailDecisions)

	srmDecision := buildSRMDecision(cfg, srm)
	srmGate := cfg.Decision.SRMGate
	if cfg.SRM != nil && cfg.SRM.GateEnabled {
		srmGate = true
	}
	srmFail := srmGate && srmDecision != nil && srmDecision.Detected

	reason := ""
	decisionValue := decision.DecisionShip
	switch {
	case primaryFail:
		decisionValue = decision.DecisionFail
		reason = "primary metric regression"
	case guardrailFail || srmFail:
		decisionValue = decision.DecisionHold
		if guardrailFail {
			reason = "guardrail regression"
		} else {
			reason = "srm detected"
		}
	}

	var win decision.Window
	if window != nil {
		win = *window
	} else {
		win = datasetWindow(exposures, outcomes)
	}

	return &decision.Artifact{
		Decision:         decisionValue,
		Reason:           reason,
		ControlVariant:   control,
		CandidateVariant: candidate,
		PrimaryMetrics:   metricDecisions,
		Guardrails:       guardrailDecisions,
		Thresholds:       thresholds,
		DatasetWindow:    win,
		Versions: decision.Versions{
			BinaryVersion: rep.BinaryVersion,
			GitCommit:     rep.GitCommit,
		},
		SRM: srmDecision,
	}
}

func buildSRMDecision(cfg ExperimentConfig, srm *report.SRMReport) *decision.SRMDecision {
	if srm == nil || cfg.SRM == nil {
		return nil
	}
	return &decision.SRMDecision{
		Detected: srm.Global.Detected,
		PValue:   srm.Global.PValue,
		Alpha:    srm.Alpha,
	}
}

func buildGuardrailDecisions(cfg GuardrailConfig, report report.GuardrailReport, candidate string) []decision.GuardrailDecision {
	decisions := []decision.GuardrailDecision{}
	for _, g := range report.PerVariant {
		if g.Variant != candidate {
			continue
		}
		decisions = append(decisions, decision.GuardrailDecision{
			Metric:    "latency_p95_ms",
			Value:     g.LatencyP95Ms,
			Threshold: cfg.MaxLatencyP95Ms,
			Passed:    g.PassLatency,
		})
		decisions = append(decisions, decision.GuardrailDecision{
			Metric:    "error_rate",
			Value:     g.ErrorRate,
			Threshold: cfg.MaxErrorRate,
			Passed:    g.PassErrorRate,
		})
		decisions = append(decisions, decision.GuardrailDecision{
			Metric:    "empty_rate",
			Value:     g.EmptyRate,
			Threshold: cfg.MaxEmptyRate,
			Passed:    g.PassEmptyRate,
		})
	}
	return decisions
}

func anyGuardrailFail(decisions []decision.GuardrailDecision) bool {
	for _, d := range decisions {
		if !d.Passed {
			return true
		}
	}
	return false
}

func selectCandidateVariant(cfg ExperimentConfig, variants []report.VariantMetrics, control string) string {
	if cfg.Decision != nil && cfg.Decision.CandidateVariant != "" {
		return cfg.Decision.CandidateVariant
	}
	candidate := ""
	max := -1
	for _, v := range variants {
		if v.Variant == control {
			continue
		}
		if v.Exposures > max {
			max = v.Exposures
			candidate = v.Variant
		}
	}
	return candidate
}

func findVariant(variants []report.VariantMetrics, name string) report.VariantMetrics {
	for _, v := range variants {
		if v.Variant == name {
			return v
		}
	}
	return report.VariantMetrics{}
}

func metricValue(metrics report.VariantMetrics, metric string) float64 {
	switch metric {
	case "ctr":
		return metrics.CTR
	case "conversion_rate":
		return metrics.ConversionRate
	case "revenue_per_request":
		return metrics.RevenuePerRequest
	default:
		return 0
	}
}

func datasetWindow(exposures []dataset.Exposure, outcomes []dataset.Outcome) decision.Window {
	start := time.Time{}
	end := time.Time{}
	update := func(ts time.Time) {
		if ts.IsZero() {
			return
		}
		if start.IsZero() || ts.Before(start) {
			start = ts
		}
		if end.IsZero() || ts.After(end) {
			end = ts
		}
	}
	for _, e := range exposures {
		update(e.Timestamp)
	}
	for _, o := range outcomes {
		update(o.Timestamp)
	}
	return decision.Window{Start: start, End: end}
}

func datasetWindowFromStream(res streamJoinResult) decision.Window {
	start := minTime(res.ExposureMin, res.OutcomeMin)
	end := maxTime(res.ExposureMax, res.OutcomeMax)
	return decision.Window{Start: start, End: end}
}

func minTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.After(b) {
		return a
	}
	return b
}

func toAccumulatorMap(variants []report.VariantMetrics) map[string]*variantAccumulator {
	out := map[string]*variantAccumulator{}
	for _, v := range variants {
		out[v.Variant] = &variantAccumulator{variant: v.Variant, exposures: v.Exposures}
	}
	return out
}

func rate(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total)
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
