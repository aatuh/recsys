package policy

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"recsys/internal/algorithm"
)

// Metrics exposes counters for policy enforcement health.
type Metrics struct {
	includeRequests *prometheus.CounterVec
	includeDropped  *prometheus.CounterVec
	includeLeak     *prometheus.CounterVec
	explicitHits    *prometheus.CounterVec
	recentHits      *prometheus.CounterVec
	ruleActions     *prometheus.CounterVec
	ruleExposure    *prometheus.CounterVec
	responseItems   *prometheus.CounterVec
	ruleZeroEffect  *prometheus.CounterVec
}

// NewMetrics registers policy metrics with the provided Prometheus registerer.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		return nil
	}

	includeRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_include_filter_requests_total",
		Help: "Count of recommendation requests specifying include tag filters",
	}, []string{"namespace", "surface"})

	includeDropped := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_include_filter_dropped_total",
		Help: "Items dropped due to include tag filters",
	}, []string{"namespace", "surface"})

	includeLeak := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_include_filter_leak_total",
		Help: "Items violating include filters that reached the final response",
	}, []string{"namespace", "surface"})

	explicitHits := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_explicit_exclude_hits_total",
		Help: "Items removed due to explicit exclude_item_ids",
	}, []string{"namespace", "surface"})

	recentHits := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_recent_event_exclude_hits_total",
		Help: "Items removed because of recent event exclusion rules",
	}, []string{"namespace", "surface"})

	ruleActions := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_rule_actions_total",
		Help: "Count of applied rule actions by type",
	}, []string{"namespace", "surface", "action"})

	ruleExposure := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_rule_exposure_total",
		Help: "Items delivered in responses due to rule actions",
	}, []string{"namespace", "surface", "action"})

	responseItems := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_response_items_total",
		Help: "Total recommendation items returned per namespace/surface",
	}, []string{"namespace", "surface"})

	ruleZeroEffect := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_rule_zero_effect_total",
		Help: "Rules that matched but produced zero exposure",
	}, []string{"namespace", "surface", "action"})

	reg.MustRegister(includeRequests, includeDropped, includeLeak, explicitHits, recentHits, ruleActions, ruleExposure, responseItems, ruleZeroEffect)

	return &Metrics{
		includeRequests: includeRequests,
		includeDropped:  includeDropped,
		includeLeak:     includeLeak,
		explicitHits:    explicitHits,
		recentHits:      recentHits,
		ruleActions:     ruleActions,
		ruleExposure:    ruleExposure,
		responseItems:   responseItems,
		ruleZeroEffect:  ruleZeroEffect,
	}
}

// Observe records metrics for a policy summary.
func (m *Metrics) Observe(req algorithm.Request, summary *algorithm.PolicySummary) {
	if m == nil || summary == nil {
		return
	}

	ns := normalizeLabel(req.Namespace, "default")
	surface := normalizeLabel(req.Surface, "default")

	if len(summary.ConstraintIncludeTags) > 0 {
		m.includeRequests.WithLabelValues(ns, surface).Inc()
		if summary.ConstraintFilteredCount > 0 {
			m.includeDropped.WithLabelValues(ns, surface).Add(float64(summary.ConstraintFilteredCount))
		}
		if summary.ConstraintLeakCount > 0 {
			m.includeLeak.WithLabelValues(ns, surface).Add(float64(summary.ConstraintLeakCount))
		}
	}

	if summary.ExplicitExcludeHits > 0 {
		m.explicitHits.WithLabelValues(ns, surface).Add(float64(summary.ExplicitExcludeHits))
	}
	if summary.RecentEventExcludeHits > 0 {
		m.recentHits.WithLabelValues(ns, surface).Add(float64(summary.RecentEventExcludeHits))
	}

	if summary.RuleBlockCount > 0 {
		m.ruleActions.WithLabelValues(ns, surface, "block").Add(float64(summary.RuleBlockCount))
	}
	if summary.RulePinCount > 0 {
		m.ruleActions.WithLabelValues(ns, surface, "pin").Add(float64(summary.RulePinCount))
	}
	if summary.RuleBoostCount > 0 {
		m.ruleActions.WithLabelValues(ns, surface, "boost").Add(float64(summary.RuleBoostCount))
	}

	if summary.FinalCount > 0 {
		m.responseItems.WithLabelValues(ns, surface).Add(float64(summary.FinalCount))
	}
	if summary.RuleBoostExposure > 0 {
		m.ruleExposure.WithLabelValues(ns, surface, "boost").Add(float64(summary.RuleBoostExposure))
	}
	if summary.RulePinExposure > 0 {
		m.ruleExposure.WithLabelValues(ns, surface, "pin").Add(float64(summary.RulePinExposure))
	}
	if summary.RuleBoostCount > 0 && summary.RuleBoostExposure == 0 {
		m.ruleZeroEffect.WithLabelValues(ns, surface, "boost").Inc()
	}
	if summary.RulePinCount > 0 && summary.RulePinExposure == 0 {
		m.ruleZeroEffect.WithLabelValues(ns, surface, "pin").Inc()
	}
}

func normalizeLabel(raw, fallback string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
