package usecase

import (
	"math"
	"sort"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

func buildExecutiveSummary(rep report.Report) *report.ExecutiveSummary {
	summary := report.ExecutiveSummary{}
	summary.Decision, summary.GateFailures = gateDecision(rep.Gates)
	summary.Highlights = pickHighlights(rep)
	summary.KeyDeltas = pickKeyDeltas(rep, 3)
	summary.NextSteps = buildNextSteps(rep, summary.Decision)
	return &summary
}

func gateDecision(gates []report.GateResult) (string, []report.GateResult) {
	if len(gates) == 0 {
		return "not_configured", nil
	}
	failures := make([]report.GateResult, 0)
	for _, g := range gates {
		if !g.Passed {
			failures = append(failures, g)
		}
	}
	if len(failures) == 0 {
		return "pass", nil
	}
	return "fail", failures
}

func pickHighlights(rep report.Report) []report.MetricResult {
	if rep.Offline != nil {
		return pickOfflineHighlights(rep.Offline.Metrics, 3)
	}
	if rep.Experiment != nil {
		return pickExperimentHighlights(rep.Experiment.Variants)
	}
	return nil
}

func pickOfflineHighlights(metrics []report.MetricResult, limit int) []report.MetricResult {
	if len(metrics) == 0 || limit <= 0 {
		return nil
	}
	priorities := []string{"ndcg", "map", "precision", "recall", "coverage", "hitrate", "mrr"}
	picked := make([]report.MetricResult, 0, limit)
	used := make(map[int]struct{})
	for _, p := range priorities {
		for i, m := range metrics {
			if _, ok := used[i]; ok {
				continue
			}
			if metricKey(m.Name) == p {
				picked = append(picked, m)
				used[i] = struct{}{}
				break
			}
		}
		if len(picked) >= limit {
			return picked
		}
	}
	for i, m := range metrics {
		if len(picked) >= limit {
			break
		}
		if _, ok := used[i]; ok {
			continue
		}
		picked = append(picked, m)
	}
	return picked
}

func pickExperimentHighlights(variants []report.VariantMetrics) []report.MetricResult {
	if len(variants) == 0 {
		return nil
	}
	idx := 0
	for i := 1; i < len(variants); i++ {
		if variants[i].Exposures > variants[idx].Exposures {
			idx = i
		}
	}
	v := variants[idx]
	return []report.MetricResult{
		{Name: "ctr", Value: v.CTR},
		{Name: "conversion_rate", Value: v.ConversionRate},
		{Name: "revenue_per_request", Value: v.RevenuePerRequest},
	}
}

func pickKeyDeltas(rep report.Report, limit int) []report.MetricDelta {
	if rep.Offline == nil || rep.Offline.Baseline == nil || len(rep.Offline.Baseline.Deltas) == 0 {
		return nil
	}
	deltas := append([]report.MetricDelta(nil), rep.Offline.Baseline.Deltas...)
	sort.SliceStable(deltas, func(i, j int) bool {
		ai := math.Abs(deltas[i].Delta)
		aj := math.Abs(deltas[j].Delta)
		if ai == aj {
			return deltas[i].Name < deltas[j].Name
		}
		return ai > aj
	})
	if limit > 0 && len(deltas) > limit {
		deltas = deltas[:limit]
	}
	return deltas
}

func buildNextSteps(rep report.Report, decision string) []string {
	steps := make([]string, 0, 3)
	if decision == "fail" {
		steps = append(steps, "Investigate failing gates and re-run after fixes.")
	}
	if rep.Offline != nil && rep.Offline.Baseline == nil {
		steps = append(steps, "Add a baseline report to measure deltas over time.")
	}
	if rep.DataQuality != nil {
		join := rep.DataQuality.JoinIntegrity
		if join.OutcomeJoinRate > 0 && join.OutcomeJoinRate < 0.8 {
			steps = append(steps, "Improve exposure/outcome joins (request_id alignment) to raise join rates.")
		}
	}
	if rep.Summary.CasesEvaluated > 0 && rep.Summary.CasesEvaluated < 50 {
		steps = append(steps, "Increase sample size to stabilize metrics.")
	}
	if len(steps) < 3 && rep.Experiment != nil && len(rep.Experiment.Variants) < 2 {
		steps = append(steps, "Include at least two variants to compare experiment outcomes.")
	}
	if len(steps) > 3 {
		steps = steps[:3]
	}
	return steps
}

func metricKey(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ""
	}
	if idx := strings.Index(name, "@"); idx > 0 {
		return name[:idx]
	}
	return name
}
