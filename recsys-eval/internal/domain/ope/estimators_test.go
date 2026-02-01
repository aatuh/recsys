package ope

import "testing"

func TestEstimators(t *testing.T) {
	samples := []Sample{
		{Reward: 1, LoggingProp: 0.5, TargetProp: 0.5, PositionWeight: 1, ModelPrediction: 0.4},
		{Reward: 0, LoggingProp: 0.5, TargetProp: 0.5, PositionWeight: 1, ModelPrediction: 0.4},
	}
	ips := IPS(samples, 0)
	if ips.Value <= 0 {
		t.Fatalf("expected ips > 0 got %.4f", ips.Value)
	}
	snips := SNIPS(samples, 0)
	if snips.Value <= 0 {
		t.Fatalf("expected snips > 0 got %.4f", snips.Value)
	}
	dr := DR(samples, 0)
	if dr.Value <= 0 {
		t.Fatalf("expected dr > 0 got %.4f", dr.Value)
	}
}
