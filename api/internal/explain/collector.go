package explain

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"recsys/internal/store"
)

// Collector assembles the compact facts pack from backing stores.
type Collector struct {
	Store *store.Store
}

// Collect gathers facts and supporting evidence.
func (c *Collector) Collect(ctx context.Context, orgID uuid.UUID, req Request) (FactsPack, []Evidence, error) {
	facts := FactsPack{}
	facts.Target.Type = strings.ToLower(req.TargetType)
	facts.Target.ID = req.TargetID
	facts.Window.From = req.From.UTC().Format(time.RFC3339)
	facts.Window.To = req.To.UTC().Format(time.RFC3339)
	facts.Context.Namespace = req.Namespace
	facts.Context.Surface = req.Surface
	if req.SegmentID != "" {
		facts.Context.SegmentID = req.SegmentID
	}
	facts.Links = map[string]string{
		"rules":   "/admin/rules?namespace=" + req.Namespace + "&surface=" + req.Surface,
		"metrics": "/dash/metrics?namespace=" + req.Namespace + "&surface=" + req.Surface,
	}

	var evidence []Evidence

	summaries, err := c.collectDecisionTraces(ctx, orgID, req)
	if err != nil {
		return facts, evidence, err
	}

	impressions := 0
	clicks := 0
	errorsCount := 0

	for _, summary := range summaries {
		// For surface targets, impressions are total items shown across
		// decisions, and clicks are any click-like reasons across items.
		// For item targets, keep legacy behavior: count only that item's
		// occurrences and click-like reasons attached to that item.
		if strings.EqualFold(req.TargetType, "surface") {
			impressions += c.countSurfaceImpressions(summary.FinalItemsJSON)
			clicks += c.countSurfaceClicks(summary.FinalItemsJSON)
		} else {
			impressions += c.countTargetImpressions(summary.FinalItemsJSON, req.TargetID)
			clicks += c.countTargetClicks(summary.FinalItemsJSON, req.TargetID)
		}

		// Augment click counts using events table where available. For surface
		// targets, count all clicks in the namespace/window. For item targets,
		// count clicks for that item id in the same window. This provides a
		// stronger signal than relying on decision-trace reasons alone.
		if strings.EqualFold(req.TargetType, "surface") {
			if evClicks, err := c.Store.CountEventsByName(ctx, orgID, req.Namespace, req.From, req.To, "", "click"); err == nil {
				clicks = evClicks
			}
		} else if req.TargetID != "" {
			if evClicks, err := c.Store.CountEventsByName(ctx, orgID, req.Namespace, req.From, req.To, req.TargetID, "click"); err == nil {
				clicks = evClicks
			}
		}
		if len(summary.ExtrasJSON) > 0 {
			c.appendAuditEvidence(summary.ExtrasJSON, &evidence)
			if c.extrasReportError(summary.ExtrasJSON) {
				errorsCount++
			}
		}
	}

	facts.Metrics.Impressions = impressions
	facts.Metrics.Clicks = clicks
	facts.Metrics.Errors = errorsCount
	if impressions > 0 {
		facts.Metrics.CTR = float64(clicks) / float64(impressions)
	}
	if impressions > 0 {
		evidence = append(evidence, Evidence{
			EvidenceID: "metric_impressions",
			Stage:      "metrics",
			Message:    "impressions=" + strconv.Itoa(impressions),
			Count:      impressions,
		})
	}
	if clicks > 0 {
		evidence = append(evidence, Evidence{
			EvidenceID: "metric_clicks",
			Stage:      "metrics",
			Message:    "clicks=" + strconv.Itoa(clicks),
			Count:      clicks,
		})
	}

	if err := c.collectRules(ctx, orgID, req, &facts, &evidence); err != nil {
		return facts, evidence, err
	}

	sort.Slice(evidence, func(i, j int) bool { return evidence[i].EvidenceID < evidence[j].EvidenceID })

	return facts, evidence, nil
}

func (c *Collector) collectDecisionTraces(ctx context.Context, orgID uuid.UUID, req Request) ([]store.DecisionTraceSummary, error) {
	from := req.From
	to := req.To
	filter := store.DecisionTraceFilter{
		From:  &from,
		To:    &to,
		Limit: 25,
	}
	return c.Store.ListDecisionTraces(ctx, orgID, req.Namespace, filter)
}

func (c *Collector) countTargetImpressions(raw []byte, targetID string) int {
	if len(raw) == 0 {
		return 0
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0
	}
	total := 0
	for _, item := range items {
		if id, _ := item["item_id"].(string); id == targetID {
			total++
		}
	}
	return total
}

func (c *Collector) countTargetClicks(raw []byte, targetID string) int {
	if len(raw) == 0 {
		return 0
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0
	}
	total := 0
	for _, item := range items {
		if id, _ := item["item_id"].(string); id != targetID {
			continue
		}
		reasons, _ := item["reasons"].([]any)
		for _, reason := range reasons {
			if s, ok := reason.(string); ok && strings.Contains(strings.ToLower(s), "click") {
				total++
			}
		}
	}
	return total
}

// countSurfaceImpressions counts total items shown in a decision's final list.
func (c *Collector) countSurfaceImpressions(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0
	}
	return len(items)
}

// countSurfaceClicks counts click-like reasons across all items in a decision.
func (c *Collector) countSurfaceClicks(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0
	}
	total := 0
	for _, item := range items {
		reasons, _ := item["reasons"].([]any)
		for _, reason := range reasons {
			if s, ok := reason.(string); ok && strings.Contains(strings.ToLower(s), "click") {
				total++
			}
		}
	}
	return total
}

func (c *Collector) appendAuditEvidence(raw []byte, evidence *[]Evidence) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	if matches, ok := payload["rules_matched"].([]any); ok {
		for _, entry := range matches {
			obj, _ := entry.(map[string]any)
			ruleID, _ := obj["rule_id"].(string)
			if ruleID == "" {
				continue
			}
			action, _ := obj["action"].(string)
			*evidence = append(*evidence, Evidence{
				EvidenceID: "rule_match:" + ruleID,
				Stage:      "rules",
				Message:    strings.ToLower(action) + " matched",
				Count:      1,
			})
		}
	}
	if effects, ok := payload["rule_effects_per_item"].(map[string]any); ok {
		for itemID, rawEffect := range effects {
			obj, _ := rawEffect.(map[string]any)
			blocked, _ := obj["blocked"].(bool)
			pinned, _ := obj["pinned"].(bool)
			boost, _ := obj["boost_delta"].(float64)
			parts := []string{"item=" + itemID}
			if blocked {
				parts = append(parts, "blocked")
			}
			if pinned {
				parts = append(parts, "pinned")
			}
			if boost != 0 {
				parts = append(parts, "boost="+formatFloat(boost))
			}
			*evidence = append(*evidence, Evidence{
				EvidenceID: "rule_effect:" + itemID,
				Stage:      "rules",
				Message:    strings.Join(parts, ", "),
				Count:      1,
			})
		}
	}
}

func (c *Collector) extrasReportError(raw []byte) bool {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	if val, ok := payload["errors"].(bool); ok {
		return val
	}
	if val, ok := payload["errors"].(float64); ok {
		return val > 0
	}
	return false
}

func (c *Collector) collectRules(ctx context.Context, orgID uuid.UUID, req Request, facts *FactsPack, evidence *[]Evidence) error {
	now := time.Now().UTC()
	rules, err := c.Store.ListActiveRulesForScope(ctx, orgID, req.Namespace, req.Surface, req.SegmentID, now)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		ttl := formatTTL(rule.ValidFrom, rule.ValidUntil)
		facts.RulesActive = append(facts.RulesActive, RuleFact{
			RuleID:   rule.RuleID.String(),
			Action:   string(rule.Action),
			Target:   string(rule.TargetType),
			Priority: rule.Priority,
			TTL:      ttl,
		})
		*evidence = append(*evidence, Evidence{
			EvidenceID: "rule_active:" + rule.RuleID.String(),
			Stage:      "rules",
			Message:    "active rule priority=" + strconv.Itoa(rule.Priority),
			Count:      1,
		})
	}
	return nil
}

func formatTTL(from, until *time.Time) string {
	switch {
	case from == nil && until == nil:
		return "open"
	case from != nil && until == nil:
		return from.UTC().Format(time.RFC3339) + ".."
	case from == nil && until != nil:
		return ".." + until.UTC().Format(time.RFC3339)
	default:
		return from.UTC().Format(time.RFC3339) + ".." + until.UTC().Format(time.RFC3339)
	}
}

func formatFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
}
