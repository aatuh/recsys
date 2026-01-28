package algorithm

import "testing"

func TestGetBlendWeightsClampsConfig(t *testing.T) {
	eng := NewEngine(Config{
		BlendAlpha: -1,
		BlendBeta:  -0.5,
		BlendGamma: -2,
	}, nil, nil)

	weights := eng.getBlendWeights(Request{})
	if weights.Pop != 1 || weights.Cooc != 0 || weights.Similarity != 0 {
		t.Fatalf("expected clamped weights {1,0,0}, got {%.2f,%.2f,%.2f}", weights.Pop, weights.Cooc, weights.Similarity)
	}
}
