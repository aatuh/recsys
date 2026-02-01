package markdown

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

// Writer writes reports as Markdown.
type Writer struct{}

func (Writer) Write(_ context.Context, rep report.Report, path string) error {
	var b strings.Builder
	b.WriteString("# Recsys Eval Report\n\n")
	b.WriteString(fmt.Sprintf("- Run ID: `%s`\n", rep.RunID))
	b.WriteString(fmt.Sprintf("- Mode: `%s`\n", rep.Mode))
	b.WriteString(fmt.Sprintf("- Created At: `%s`\n", rep.CreatedAt.Format("2006-01-02T15:04:05Z")) + "\n")
	if rep.Summary.CasesEvaluated > 0 {
		b.WriteString(fmt.Sprintf("- Cases Evaluated: `%d`\n\n", rep.Summary.CasesEvaluated))
	} else {
		b.WriteString("\n")
	}

	if rep.Offline != nil {
		b.WriteString("## Offline Metrics\n\n")
		writeMetricTable(&b, rep.Offline.Metrics)
	}

	if rep.Experiment != nil {
		b.WriteString("\n## Experiment Metrics\n\n")
		writeVariantTable(&b, rep.Experiment.Variants)
		if len(rep.Experiment.Guardrails.PerVariant) > 0 {
			b.WriteString("\n### Guardrails\n\n")
			writeGuardrailTable(&b, rep.Experiment.Guardrails.PerVariant)
		}
	}

	if len(rep.Gates) > 0 {
		b.WriteString("\n## Regression Gates\n\n")
		writeGateTable(&b, rep.Gates)
	}

	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func writeMetricTable(b *strings.Builder, metrics []report.MetricResult) {
	if len(metrics) == 0 {
		b.WriteString("_No metrics computed._\n")
		return
	}
	b.WriteString("| Metric | Value | CI |\n| --- | --- | --- |\n")
	for _, m := range metrics {
		ci := ""
		if m.CI != nil {
			ci = fmt.Sprintf("[%.4f, %.4f] (%.2f)", m.CI.Lower, m.CI.Upper, m.CI.Level)
		}
		fmt.Fprintf(b, "| %s | %.6f | %s |\n", m.Name, m.Value, ci)
	}
}

func writeVariantTable(b *strings.Builder, variants []report.VariantMetrics) {
	if len(variants) == 0 {
		b.WriteString("_No variants computed._\n")
		return
	}
	b.WriteString("| Variant | Exposures | CTR | CVR | Revenue/Req | Throughput RPS |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, v := range variants {
		fmt.Fprintf(
			b,
			"| %s | %d | %.4f | %.4f | %.4f | %.4f |\n",
			v.Variant, v.Exposures, v.CTR, v.ConversionRate, v.RevenuePerRequest, v.ThroughputRPS,
		)
	}
}

func writeGuardrailTable(b *strings.Builder, metrics []report.GuardrailMetrics) {
	b.WriteString("| Variant | Latency p95 (ms) | Error Rate | Empty Rate | Pass |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, m := range metrics {
		pass := m.PassLatency && m.PassErrorRate && m.PassEmptyRate
		fmt.Fprintf(
			b,
			"| %s | %.2f | %.4f | %.4f | %t |\n",
			m.Variant, m.LatencyP95Ms, m.ErrorRate, m.EmptyRate, pass,
		)
	}
}

func writeGateTable(b *strings.Builder, gates []report.GateResult) {
	b.WriteString("| Metric | Delta | Max Drop | Passed |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, g := range gates {
		fmt.Fprintf(b, "| %s | %.6f | %.6f | %t |\n", g.Metric, g.Delta, g.MaxDrop, g.Passed)
	}
}
