package statistics

import "math"

// TwoProportionZTestPValue returns the two-sided p-value for difference in proportions.
func TwoProportionZTestPValue(success1, total1, success0, total0 int) float64 {
	if total1 == 0 || total0 == 0 {
		return 1
	}
	p1 := float64(success1) / float64(total1)
	p0 := float64(success0) / float64(total0)
	p := float64(success1+success0) / float64(total1+total0)
	se := math.Sqrt(p * (1 - p) * (1/float64(total1) + 1/float64(total0)))
	if se == 0 {
		return 1
	}
	z := (p1 - p0) / se
	pval := 2 * (1 - NormalCDF(math.Abs(z)))
	if pval < 0 {
		return 0
	}
	if pval > 1 {
		return 1
	}
	return pval
}
