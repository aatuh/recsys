package ranking

import "github.com/aatuh/recsys-suite/recsys-eval/internal/domain/metrics"

// RegisterDefaults adds the standard ranking metrics to the registry.
func RegisterDefaults(reg *metrics.Registry) {
	reg.Register("precision", NewPrecisionAtK)
	reg.Register("precision@k", NewPrecisionAtK)
	reg.Register("recall", NewRecallAtK)
	reg.Register("recall@k", NewRecallAtK)
	reg.Register("map", NewMAPAtK)
	reg.Register("map@k", NewMAPAtK)
	reg.Register("ndcg", NewNDCGAtK)
	reg.Register("ndcg@k", NewNDCGAtK)
	reg.Register("hitrate", NewHitRateAtK)
	reg.Register("hitrate@k", NewHitRateAtK)
}
