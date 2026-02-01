package html

import (
	"context"
	"html/template"
	"os"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

// Writer writes reports as a simple HTML page.
type Writer struct{}

type viewModel struct {
	Report report.Report
}

const htmlTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Recsys Eval Report</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 2rem; }
    table { border-collapse: collapse; width: 100%; margin-bottom: 2rem; }
    th, td { border: 1px solid #ddd; padding: 0.5rem; text-align: left; }
    th { background: #f6f6f6; }
    code { background: #f2f2f2; padding: 0.1rem 0.25rem; }
  </style>
</head>
<body>
  <h1>Recsys Eval Report</h1>
  <p><strong>Run ID:</strong> <code>{{ .Report.RunID }}</code></p>
  <p><strong>Mode:</strong> <code>{{ .Report.Mode }}</code></p>
  <p><strong>Created At:</strong> <code>{{ .Report.CreatedAt.Format "2006-01-02T15:04:05Z" }}</code></p>
  {{ if gt .Report.Summary.CasesEvaluated 0 }}
  <p><strong>Cases Evaluated:</strong> <code>{{ .Report.Summary.CasesEvaluated }}</code></p>
  {{ end }}

  {{ if .Report.Summary.Executive }}
  <h2>Executive Summary</h2>
  {{ if .Report.Summary.Executive.Decision }}
  <p><strong>Decision:</strong> <code>{{ .Report.Summary.Executive.Decision }}</code></p>
  {{ end }}
  {{ if .Report.Summary.Executive.Highlights }}
  <h3>Highlights</h3>
  <ul>
    {{ range .Report.Summary.Executive.Highlights }}
    <li>{{ .Name }}: {{ printf "%.6f" .Value }}</li>
    {{ end }}
  </ul>
  {{ end }}
  {{ if .Report.Summary.Executive.KeyDeltas }}
  <h3>Key Deltas vs Baseline</h3>
  <ul>
    {{ range .Report.Summary.Executive.KeyDeltas }}
    <li>{{ .Name }}: {{ printf "%.6f" .Delta }}</li>
    {{ end }}
  </ul>
  {{ end }}
  {{ if .Report.Summary.Executive.NextSteps }}
  <h3>Next Steps</h3>
  <ul>
    {{ range .Report.Summary.Executive.NextSteps }}
    <li>{{ . }}</li>
    {{ end }}
  </ul>
  {{ end }}
  {{ end }}

  {{ if .Report.Offline }}
  <h2>Offline Metrics</h2>
  {{ if .Report.Offline.Metrics }}
  <table>
    <thead><tr><th>Metric</th><th>Value</th><th>CI</th></tr></thead>
    <tbody>
      {{ range .Report.Offline.Metrics }}
      <tr>
        <td>{{ .Name }}</td>
        <td>{{ printf "%.6f" .Value }}</td>
        <td>{{ if .CI }}[{{ printf "%.4f" .CI.Lower }}, {{ printf "%.4f" .CI.Upper }}] ({{ printf "%.2f" .CI.Level }}){{ end }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ else }}
  <p>No metrics computed.</p>
  {{ end }}
  {{ end }}

  {{ if .Report.Experiment }}
  <h2>Experiment Metrics</h2>
  {{ if .Report.Experiment.Variants }}
  <table>
    <thead><tr><th>Variant</th><th>Exposures</th><th>CTR</th><th>CVR</th><th>Revenue/Req</th><th>Throughput RPS</th></tr></thead>
    <tbody>
      {{ range .Report.Experiment.Variants }}
      <tr>
        <td>{{ .Variant }}</td>
        <td>{{ .Exposures }}</td>
        <td>{{ printf "%.4f" .CTR }}</td>
        <td>{{ printf "%.4f" .ConversionRate }}</td>
        <td>{{ printf "%.4f" .RevenuePerRequest }}</td>
        <td>{{ printf "%.4f" .ThroughputRPS }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ end }}

  {{ if .Report.Experiment.Guardrails.PerVariant }}
  <h3>Guardrails</h3>
  <table>
    <thead><tr><th>Variant</th><th>Latency p95 (ms)</th><th>Error Rate</th><th>Empty Rate</th><th>Pass</th></tr></thead>
    <tbody>
      {{ range .Report.Experiment.Guardrails.PerVariant }}
      <tr>
        <td>{{ .Variant }}</td>
        <td>{{ printf "%.2f" .LatencyP95Ms }}</td>
        <td>{{ printf "%.4f" .ErrorRate }}</td>
        <td>{{ printf "%.4f" .EmptyRate }}</td>
        <td>{{ and .PassLatency .PassErrorRate .PassEmptyRate }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ end }}
  {{ end }}

  {{ if .Report.Gates }}
  <h2>Regression Gates</h2>
  <table>
    <thead><tr><th>Metric</th><th>Delta</th><th>Max Drop</th><th>Passed</th></tr></thead>
    <tbody>
      {{ range .Report.Gates }}
      <tr>
        <td>{{ .Metric }}</td>
        <td>{{ printf "%.6f" .Delta }}</td>
        <td>{{ printf "%.6f" .MaxDrop }}</td>
        <td>{{ .Passed }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ end }}
</body>
</html>`

func (Writer) Write(_ context.Context, rep report.Report, path string) error {
	tpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return tpl.Execute(file, viewModel{Report: rep})
}
