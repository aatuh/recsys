package algorithm

import "testing"

func TestNormalizeEmbeddingScoreClamps(t *testing.T) {
	cases := []struct {
		input float64
		want  float64
	}{
		{input: -1, want: 0},
		{input: 0, want: 0},
		{input: 0.25, want: 0.25},
		{input: 1, want: 1},
		{input: 1.5, want: 1},
	}

	for _, tc := range cases {
		if got := normalizeEmbeddingScore(tc.input); got != tc.want {
			t.Fatalf("normalizeEmbeddingScore(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestNormalizePositiveScoreMonotonic(t *testing.T) {
	positives := []float64{0.1, 0.5, 1, 2, 10}
	prev := 0.0
	for i, v := range positives {
		got := normalizePositiveScore(v)
		if got <= 0 || got >= 1 {
			t.Fatalf("normalizePositiveScore(%v) = %v, want in (0,1)", v, got)
		}
		if i > 0 && got < prev {
			t.Fatalf("normalizePositiveScore(%v) = %v, want >= %v", v, got, prev)
		}
		prev = got
	}

	if got := normalizePositiveScore(0); got != 0 {
		t.Fatalf("normalizePositiveScore(0) = %v, want 0", got)
	}
	if got := normalizePositiveScore(-2); got != 0 {
		t.Fatalf("normalizePositiveScore(-2) = %v, want 0", got)
	}
}
