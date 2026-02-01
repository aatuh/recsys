package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/ope"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/clock"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/reporting"
)

// OPEUsecase orchestrates off-policy evaluation from logs.
type OPEUsecase struct {
	Exposures datasource.ExposureReader
	Outcomes  datasource.OutcomeReader
	Reporter  reporting.Writer
	Clock     clock.Clock
	Logger    logger.Logger
	Metadata  ReportMetadata
	Scale     ScaleConfig
}

func (u OPEUsecase) Run(ctx context.Context, evalCfg OPEConfig, outputPath string) (report.Report, error) {
	if u.Exposures == nil || u.Outcomes == nil {
		return report.Report{}, errors.New("exposure and outcome readers are required")
	}

	exposures, err := u.Exposures.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}
	outcomes, err := u.Outcomes.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}

	rewardEvent := strings.ToLower(evalCfg.RewardEvent)
	if rewardEvent == "" {
		rewardEvent = "click"
	}
	if rewardEvent != "click" && rewardEvent != "conversion" {
		return report.Report{}, fmt.Errorf("unsupported reward_event: %s", evalCfg.RewardEvent)
	}

	outcomeByReq := map[string][]dataset.Outcome{}
	for _, o := range outcomes {
		outcomeByReq[o.RequestID] = append(outcomeByReq[o.RequestID], o)
	}
	_, joinStats := dataset.JoinByRequest(exposures, outcomes)

	// Build reward model per item for DR.
	itemRewardSum := map[string]float64{}
	itemRewardCount := map[string]int{}
	for _, o := range outcomes {
		if strings.ToLower(o.EventType) != rewardEvent {
			continue
		}
		reward := rewardValue(o)
		itemRewardSum[o.ItemID] += reward
		itemRewardCount[o.ItemID]++
	}

	maxK := evalCfg.TopK
	if maxK <= 0 {
		maxK = math.MaxInt
	}

	minProp := evalCfg.MinPropensity
	if minProp <= 0 {
		minProp = 1e-6
	}

	unit := strings.ToLower(evalCfg.Unit)
	if unit == "" {
		unit = "request"
	}
	rewardAgg := strings.ToLower(evalCfg.RewardAggregation)
	if rewardAgg == "" {
		rewardAgg = "sum"
	}

	var samples []ope.Sample
	weights := make([]float64, 0)
	missingTarget := 0
	nearZero := 0
	clipped := 0

	for _, exp := range exposures {
		items := append([]dataset.ExposedItem(nil), exp.Items...)
		sort.SliceStable(items, func(i, j int) bool { return items[i].Rank < items[j].Rank })
		if len(items) > maxK {
			items = items[:maxK]
		}

		outcomes := outcomeByReq[exp.RequestID]
		outByItem := map[string]dataset.Outcome{}
		for _, o := range outcomes {
			outByItem[o.ItemID] = o
		}

		reqWeights := make([]float64, 0, len(items))
		reqRewards := make([]float64, 0, len(items))
		reqPreds := make([]float64, 0, len(items))

		for _, item := range items {
			logProp, ok := loggingPropensity(item)
			if !ok || logProp <= 0 {
				continue
			}
			if logProp < minProp {
				nearZero++
			}

			targetProp, ok := targetPropensity(item)
			if !ok {
				if !evalCfg.AllowMissingTargetPropensity {
					missingTarget++
					continue
				}
				targetProp = logProp
				missingTarget++
			}
			if targetProp < minProp {
				nearZero++
			}

			posWeight := positionWeight(item.Rank, evalCfg.PositionDiscount)
			reward := 0.0
			if out, ok := outByItem[item.ItemID]; ok {
				if strings.ToLower(out.EventType) == rewardEvent {
					reward = rewardValue(out)
				}
			}

			model := 0.0
			if cnt := itemRewardCount[item.ItemID]; cnt > 0 {
				model = itemRewardSum[item.ItemID] / float64(cnt)
			}

			w := targetProp / logProp
			if evalCfg.Clipping > 0 && w > evalCfg.Clipping {
				clipped++
				w = evalCfg.Clipping
			}
			if unit == "item" {
				sample := ope.Sample{
					Reward:          reward,
					LoggingProp:     logProp,
					TargetProp:      targetProp,
					PositionWeight:  posWeight,
					ModelPrediction: model,
				}
				weights = append(weights, w)
				samples = append(samples, sample)
				continue
			}

			reqWeights = append(reqWeights, w)
			reqRewards = append(reqRewards, reward*posWeight)
			reqPreds = append(reqPreds, model*posWeight)
		}

		if unit == "request" && len(reqWeights) > 0 {
			reqWeight := mean(reqWeights)
			reqReward := aggregate(reqRewards, rewardAgg)
			reqPred := aggregate(reqPreds, rewardAgg)
			sample := ope.Sample{
				Reward:          reqReward,
				LoggingProp:     1,
				TargetProp:      reqWeight,
				PositionWeight:  1,
				ModelPrediction: reqPred,
			}
			weights = append(weights, reqWeight)
			samples = append(samples, sample)
		}
	}

	if len(samples) == 0 {
		return report.Report{}, fmt.Errorf("no valid samples for OPE")
	}

	results := []report.OPEEstimatorResult{}
	ips := ope.IPS(samples, evalCfg.Clipping)
	results = append(results, report.OPEEstimatorResult{Name: "ips", Value: ips.Value, Variance: ips.Variance})
	if evalCfg.EnableSNIPS {
		snips := ope.SNIPS(samples, evalCfg.Clipping)
		results = append(results, report.OPEEstimatorResult{Name: "snips", Value: snips.Value, Variance: snips.Variance})
	}
	if evalCfg.EnableDR {
		dr := ope.DR(samples, evalCfg.Clipping)
		results = append(results, report.OPEEstimatorResult{Name: "dr", Value: dr.Value, Variance: dr.Variance})
	}

	ess := ope.EffectiveSampleSize(weights)
	stats := report.WeightStats{
		Count:               len(weights),
		Mean:                mean(weights),
		P95:                 ope.Percentile(weights, 0.95),
		Max:                 max(weights),
		EffectiveSampleSize: ess,
		ClippedFraction:     fraction(clipped, len(weights)),
	}

	oped := report.OPEDiagnostics{
		MissingTargetPropensity: missingTarget,
		NearZeroPropensity:      nearZero,
		TotalItems:              len(weights),
	}

	warnings := []string{}
	if stats.ClippedFraction > 0.1 {
		warnings = append(warnings, "high_clipping_fraction")
	}
	if ess > 0 && ess < float64(len(weights))*0.2 {
		warnings = append(warnings, "low_effective_sample_size")
	}
	if missingTarget > 0 && !evalCfg.AllowMissingTargetPropensity {
		warnings = append(warnings, "missing_target_propensity")
	}

	rep := report.Report{
		RunID:                   fmt.Sprintf("ope-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    "ope",
		CreatedAt:               u.Clock.Now().UTC(),
		Version:                 "0.1.0",
		BinaryVersion:           u.Metadata.BinaryVersion,
		GitCommit:               u.Metadata.GitCommit,
		EffectiveConfig:         u.Metadata.EffectiveConfig,
		InputDatasetFingerprint: u.Metadata.InputDatasetFingerprint,
		Artifacts:               u.Metadata.Artifacts,
		Summary: report.Summary{
			CasesEvaluated: len(samples),
		},
		OPE: &report.OPEReport{
			Unit:              unit,
			RewardAggregation: rewardAgg,
			Estimators:        results,
			WeightStats:       stats,
			Diagnostics:       oped,
		},
		DataQuality: func() *report.DataQualityReport {
			dq := buildDataQuality(exposures, outcomes, joinStats, 0)
			return &dq
		}(),
		Warnings: warnings,
	}
	rep.Summary.Executive = buildExecutiveSummary(rep)

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, err
	}

	return rep, nil
}

func rewardValue(o dataset.Outcome) float64 {
	if o.Value != 0 {
		return o.Value
	}
	return 1
}

func loggingPropensity(item dataset.ExposedItem) (float64, bool) {
	if item.LoggingPropensity != nil {
		return *item.LoggingPropensity, true
	}
	if item.Propensity != nil {
		return *item.Propensity, true
	}
	return 0, false
}

func targetPropensity(item dataset.ExposedItem) (float64, bool) {
	if item.TargetPropensity != nil {
		return *item.TargetPropensity, true
	}
	return 0, false
}

func positionWeight(rank int, mode string) float64 {
	switch strings.ToLower(mode) {
	case "log2":
		return 1.0 / math.Log2(float64(rank)+1)
	default:
		return 1
	}
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func fraction(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func aggregate(values []float64, mode string) float64 {
	if len(values) == 0 {
		return 0
	}
	if mode == "max" {
		maxVal := values[0]
		for _, v := range values[1:] {
			if v > maxVal {
				maxVal = v
			}
		}
		return maxVal
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}
