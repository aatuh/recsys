package validation

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aatuh/recsys-suite/api/src/specs/types"
)

const (
	codeInvalidConfig  = "RECSYS_INVALID_CONFIG"
	codeInvalidRules   = "RECSYS_INVALID_RULES"
	codeInvalidTargets = "RECSYS_INVALID_TARGETS"
)

type configPayload struct {
	Weights *types.Weights  `json:"weights,omitempty"`
	Flags   map[string]bool `json:"flags,omitempty"`
	Limits  *configLimits   `json:"limits,omitempty"`
}

type configLimits struct {
	MaxK          *int `json:"max_k,omitempty"`
	MaxExcludeIDs *int `json:"max_exclude_ids,omitempty"`
}

// ValidateConfigPayload validates a tenant config document.
func ValidateConfigPayload(raw []byte) error {
	if len(bytesTrim(raw)) == 0 {
		return newError(http.StatusBadRequest, codeInvalidConfig, "config payload is required")
	}
	var payload configPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return newError(http.StatusBadRequest, codeInvalidConfig, "config payload must be valid json")
	}
	if payload.Weights != nil {
		if payload.Weights.Pop < 0 || payload.Weights.Cooc < 0 || payload.Weights.Emb < 0 {
			return newError(http.StatusUnprocessableEntity, codeInvalidConfig, "weights must be non-negative")
		}
	}
	if payload.Limits != nil {
		if payload.Limits.MaxK != nil && *payload.Limits.MaxK < 0 {
			return newError(http.StatusUnprocessableEntity, codeInvalidConfig, "limits.max_k must be non-negative")
		}
		if payload.Limits.MaxExcludeIDs != nil && *payload.Limits.MaxExcludeIDs < 0 {
			return newError(http.StatusUnprocessableEntity, codeInvalidConfig, "limits.max_exclude_ids must be non-negative")
		}
	}
	return nil
}

// ValidateRulesPayload validates a tenant rules document.
func ValidateRulesPayload(raw []byte) error {
	if len(bytesTrim(raw)) == 0 {
		return newError(http.StatusBadRequest, codeInvalidRules, "rules payload is required")
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return newError(http.StatusBadRequest, codeInvalidRules, "rules payload must be valid json")
	}
	switch payload.(type) {
	case []any, map[string]any:
		return nil
	default:
		return newError(http.StatusUnprocessableEntity, codeInvalidRules, "rules payload must be a json object or array")
	}
}

// ValidateCacheInvalidate validates cache invalidation request.
func ValidateCacheInvalidate(req *types.CacheInvalidateRequest) error {
	if req == nil {
		return newError(http.StatusBadRequest, codeInvalidTargets, "targets are required")
	}
	targets := normalizeList(req.Targets, true)
	if len(targets) == 0 {
		return newError(http.StatusBadRequest, codeInvalidTargets, "targets are required")
	}
	for _, target := range targets {
		switch target {
		case "rules", "config", "popularity":
			continue
		default:
			return newError(http.StatusUnprocessableEntity, codeInvalidTargets, "invalid cache invalidation target: "+target)
		}
	}
	return nil
}

func bytesTrim(raw []byte) []byte {
	return []byte(strings.TrimSpace(string(raw)))
}
