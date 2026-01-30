package exposure

import (
	"errors"
	"strings"
	"time"
)

const (
	LogFormatServiceV1 = "service_v1"
	LogFormatEvalV1    = "eval_v1"
)

type evalExposure struct {
	RequestID string             `json:"request_id"`
	UserID    string             `json:"user_id"`
	Timestamp time.Time          `json:"ts"`
	Items     []evalExposureItem `json:"items"`
	Context   map[string]string  `json:"context,omitempty"`
	LatencyMs *float64           `json:"latency_ms,omitempty"`
	Error     *bool              `json:"error,omitempty"`
}

type evalExposureItem struct {
	ItemID            string   `json:"item_id"`
	Rank              int      `json:"rank"`
	Propensity        *float64 `json:"propensity,omitempty"`
	LoggingPropensity *float64 `json:"logging_propensity,omitempty"`
	TargetPropensity  *float64 `json:"target_propensity,omitempty"`
}

func normalizeLogFormat(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "eval", "eval_v1":
		return LogFormatEvalV1
	case "service", "service_v1":
		return LogFormatServiceV1
	default:
		return ""
	}
}

func buildEvalExposure(event Event) (evalExposure, error) {
	requestID := strings.TrimSpace(event.RequestID)
	if requestID == "" {
		return evalExposure{}, errors.New("request_id is required for eval exposure")
	}
	userID := selectEvalUserID(event)
	if userID == "" {
		userID = requestID
	}
	ts := event.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	items := make([]evalExposureItem, 0, len(event.Items))
	for _, item := range event.Items {
		items = append(items, evalExposureItem{
			ItemID: item.ItemID,
			Rank:   item.Rank,
		})
	}
	return evalExposure{
		RequestID: requestID,
		UserID:    userID,
		Timestamp: ts,
		Items:     items,
		Context:   buildEvalContext(event),
	}, nil
}

func selectEvalUserID(event Event) string {
	if event.Subject == nil {
		return ""
	}
	if v := strings.TrimSpace(event.Subject.UserIDHash); v != "" {
		return v
	}
	if v := strings.TrimSpace(event.Subject.AnonymousIDHash); v != "" {
		return v
	}
	return strings.TrimSpace(event.Subject.SessionIDHash)
}

func buildEvalContext(event Event) map[string]string {
	ctx := map[string]string{}
	if event.TenantID != "" {
		ctx["tenant_id"] = event.TenantID
	}
	if event.Surface != "" {
		ctx["surface"] = event.Surface
	}
	if event.Segment != "" {
		ctx["segment"] = event.Segment
	}
	if event.AlgoVersion != "" {
		ctx["algo_version"] = event.AlgoVersion
	}
	if event.ConfigVersion != "" {
		ctx["config_version"] = event.ConfigVersion
	}
	if event.RulesVersion != "" {
		ctx["rules_version"] = event.RulesVersion
	}
	if event.Experiment != nil {
		if v := strings.TrimSpace(event.Experiment.ID); v != "" {
			ctx["experiment_id"] = v
		}
		if v := strings.TrimSpace(event.Experiment.Variant); v != "" {
			ctx["experiment_variant"] = v
		}
	}
	if event.Context != nil {
		if v := strings.TrimSpace(event.Context.Locale); v != "" {
			ctx["locale"] = v
		}
		if v := strings.TrimSpace(event.Context.Device); v != "" {
			ctx["device"] = v
		}
		if v := strings.TrimSpace(event.Context.Country); v != "" {
			ctx["country"] = v
		}
		if v := strings.TrimSpace(event.Context.Now); v != "" {
			ctx["now"] = v
		}
	}
	if len(ctx) == 0 {
		return nil
	}
	return ctx
}
