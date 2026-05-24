package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestRecordRecommendationMetrics(t *testing.T) {
	RecordRecommendation("recommend", "success", 0, 2, 0.015)
	RecordRecommendation("recommend", "overload", 0, 0, 0.001)

	if got := counterValue(t, "recsys_recommendation_requests_total", map[string]string{
		"endpoint": "recommend",
		"outcome":  "success",
		"empty":    "true",
	}); got < 1 {
		t.Fatalf("success empty counter = %v, want at least 1", got)
	}
	if got := counterValue(t, "recsys_recommendation_requests_total", map[string]string{
		"endpoint": "recommend",
		"outcome":  "overload",
		"empty":    "false",
	}); got < 1 {
		t.Fatalf("overload counter = %v, want at least 1", got)
	}
}

func TestArtifactMetrics(t *testing.T) {
	RecordArtifactLoadFailure("manifest")
	RecordArtifactManifest(42, 5)

	if got := counterValue(t, "recsys_artifact_load_failures_total", map[string]string{"kind": "manifest"}); got < 1 {
		t.Fatalf("artifact failure counter = %v, want at least 1", got)
	}
	if got := gaugeValue(t, "recsys_artifact_manifest_age_seconds", nil); got != 42 {
		t.Fatalf("manifest age gauge = %v, want 42", got)
	}
	if got := gaugeValue(t, "recsys_artifact_manifest_artifacts", nil); got != 5 {
		t.Fatalf("manifest artifacts gauge = %v, want 5", got)
	}
}

func counterValue(t *testing.T, name string, labels map[string]string) float64 {
	t.Helper()
	metric := findMetric(t, name, labels)
	if metric.GetCounter() == nil {
		t.Fatalf("%s is not a counter", name)
	}
	return metric.GetCounter().GetValue()
}

func gaugeValue(t *testing.T, name string, labels map[string]string) float64 {
	t.Helper()
	metric := findMetric(t, name, labels)
	if metric.GetGauge() == nil {
		t.Fatalf("%s is not a gauge", name)
	}
	return metric.GetGauge().GetValue()
}

func findMetric(t *testing.T, name string, labels map[string]string) *dto.Metric {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			if labelsMatch(metric, labels) {
				return metric
			}
		}
	}
	t.Fatalf("metric %s with labels %+v not found", name, labels)
	return nil
}

func labelsMatch(metric *dto.Metric, want map[string]string) bool {
	for k, v := range want {
		found := false
		for _, label := range metric.GetLabel() {
			if label.GetName() == k && label.GetValue() == v {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
