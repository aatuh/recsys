package statistics

import "testing"

func TestBootstrapMean(t *testing.T) {
	samples := []float64{0, 1, 0, 1, 1}
	res := BootstrapMean(samples, 200, 42, 0.95)
	if res == nil {
		t.Fatalf("expected bootstrap result")
	}
	if res.Lower > res.Upper {
		t.Fatalf("invalid CI bounds: %.4f > %.4f", res.Lower, res.Upper)
	}
	if res.Lower < 0 || res.Upper > 1 {
		t.Fatalf("CI bounds out of expected range: %.4f %.4f", res.Lower, res.Upper)
	}
}
