package policy

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aatuh/recsys-algo/rules"
	"recsys/internal/types"

	"github.com/aatuh/recsys-algo/algorithm"
)

// Metrics exposes counters for policy enforcement health.
type Metrics struct {
	includeRequests  *prometheus.CounterVec
	includeDropped   *prometheus.CounterVec
	includeLeak      *prometheus.CounterVec
	explicitHits     *prometheus.CounterVec
	recentHits       *prometheus.CounterVec
	ruleActions      *prometheus.CounterVec
	ruleExposure     *prometheus.CounterVec
	responseItems    *prometheus.CounterVec
	ruleZeroEffect   *prometheus.CounterVec
	itemExposure     *prometheus.CounterVec
	coverageBucket   *prometheus.CounterVec
	catalogSize      *prometheus.GaugeVec
	overrideMatches  *prometheus.CounterVec
	overrideExposure *prometheus.CounterVec
	constraintLeak   *prometheus.CounterVec
	ruleBlockedItems *prometheus.CounterVec
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

	itemExposure := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_item_served_total",
		Help: "Recommendation items served to clients by namespace, surface, and item_id",
	}, []string{"namespace", "surface", "item_id"})

	overrideMatches := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_override_matches_total",
		Help: "Count of manual override matches per override_id and action",
	}, []string{"namespace", "surface", "override_id", "action"})

	overrideExposure := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_override_exposure_total",
		Help: "Number of items affected by manual overrides per override_id and action",
	}, []string{"namespace", "surface", "override_id", "action"})

	constraintLeak := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_constraint_leak_total",
		Help: "Items that bypassed constraint filters and still surfaced, bucketed by constraint type",
	}, []string{"namespace", "surface", "reason"})

	ruleBlockedItems := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_rule_blocked_items_total",
		Help: "Items removed from responses due to block rules, labeled per rule_id",
	}, []string{"namespace", "surface", "rule_id"})

	coverageBucket := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_coverage_bucket_total",
		Help: "Recommendation items served grouped by coverage bucket",
	}, []string{"namespace", "surface", "bucket"})

	catalogSize := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_catalog_items_total",
		Help: "Available catalog size per namespace for coverage guardrails",
	}, []string{"namespace"})

	reg.MustRegister(includeRequests, includeDropped, includeLeak, explicitHits, recentHits, ruleActions, ruleExposure, responseItems, ruleZeroEffect, itemExposure, coverageBucket, catalogSize, overrideMatches, overrideExposure, constraintLeak, ruleBlockedItems)

	return &Metrics{
		includeRequests:  includeRequests,
		includeDropped:   includeDropped,
		includeLeak:      includeLeak,
		explicitHits:     explicitHits,
		recentHits:       recentHits,
		ruleActions:      ruleActions,
		ruleExposure:     ruleExposure,
		responseItems:    responseItems,
		ruleZeroEffect:   ruleZeroEffect,
		itemExposure:     itemExposure,
		coverageBucket:   coverageBucket,
		catalogSize:      catalogSize,
		overrideMatches:  overrideMatches,
		overrideExposure: overrideExposure,
		constraintLeak:   constraintLeak,
		ruleBlockedItems: ruleBlockedItems,
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
	if len(summary.RuleBlockExposureByRule) > 0 {
		for ruleID, count := range summary.RuleBlockExposureByRule {
			label := normalizeLabel(ruleID, "unknown")
			m.ruleBlockedItems.WithLabelValues(ns, surface, label).Add(float64(count))
		}
	} else if summary.RuleBlockExposure > 0 {
		m.ruleBlockedItems.WithLabelValues(ns, surface, "unknown").Add(float64(summary.RuleBlockExposure))
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
	if len(summary.ConstraintLeakByReason) > 0 {
		for reason, count := range summary.ConstraintLeakByReason {
			label := normalizeLabel(reason, "unknown")
			m.constraintLeak.WithLabelValues(ns, surface, label).Add(float64(count))
		}
	} else if summary.ConstraintLeakCount > 0 {
		m.constraintLeak.WithLabelValues(ns, surface, "unknown").Add(float64(summary.ConstraintLeakCount))
	}
}

// ObserveOverrides records per-override counters for manual override activity.
func (m *Metrics) ObserveOverrides(req algorithm.Request, hits []rules.OverrideHit) {
	if m == nil || len(hits) == 0 {
		return
	}
	ns := normalizeLabel(req.Namespace, "default")
	surface := normalizeLabel(req.Surface, "default")

	for _, hit := range hits {
		action := strings.ToLower(string(hit.Action))
		overrideID := hit.OverrideID.String()
		if len(hit.MatchedItems) > 0 {
			m.overrideMatches.WithLabelValues(ns, surface, overrideID, action).Add(float64(len(hit.MatchedItems)))
		}
		var exposure int
		switch hit.Action {
		case types.RuleActionBlock:
			exposure = len(hit.BlockedItems)
		case types.RuleActionPin:
			exposure = len(hit.PinnedItems)
		case types.RuleActionBoost:
			exposure = len(hit.ServedItems)
		default:
			exposure = len(hit.ServedItems)
		}
		if exposure > 0 {
			m.overrideExposure.WithLabelValues(ns, surface, overrideID, action).Add(float64(exposure))
		}
	}
}

// ObserveCoverage records item-level coverage telemetry for guardrails.
func (m *Metrics) ObserveCoverage(req algorithm.Request, itemIDs []string, longTail []bool, catalogTotal int) {
	if m == nil {
		return
	}
	ns := normalizeLabel(req.Namespace, "default")
	surface := normalizeLabel(req.Surface, "default")

	if catalogTotal >= 0 {
		m.catalogSize.WithLabelValues(ns).Set(float64(catalogTotal))
	}

	totalItems := len(itemIDs)
	for idx, id := range itemIDs {
		if id == "" {
			continue
		}
		m.itemExposure.WithLabelValues(ns, surface, id).Inc()
		m.coverageBucket.WithLabelValues(ns, surface, "all").Inc()
		if idx < len(longTail) && longTail[idx] {
			m.coverageBucket.WithLabelValues(ns, surface, "long_tail").Inc()
		}
	}

	// Ensure the bucket metric advances even when no items are returned.
	if totalItems == 0 {
		m.coverageBucket.WithLabelValues(ns, surface, "all").Add(0)
		m.coverageBucket.WithLabelValues(ns, surface, "long_tail").Add(0)
	}
}

func normalizeLabel(raw, fallback string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
