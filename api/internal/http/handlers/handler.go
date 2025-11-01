package handlers

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"recsys/internal/audit"
	"recsys/internal/explain"
	"recsys/internal/rules"
	"recsys/internal/store"
	internaltypes "recsys/internal/types"
)

// Handler retains the legacy all-in-one adapter while specific domains migrate
// to focused structs. New endpoints should prefer dedicated handler types.
type Handler struct {
	Store                 *store.Store
	DefaultOrg            uuid.UUID
	HalfLifeDays          float64
	CoVisWindowDays       float64
	PopularityFanout      int
	MMRLambda             float64
	BrandCap              int
	CategoryCap           int
	RuleExcludeEvents     bool
	ExcludeEventTypes     []int16
	BrandTagPrefixes      []string
	CategoryTagPrefixes   []string
	RulesManager          *rules.Manager
	RulesAuditSample      float64
	ExplainService        *explain.Service
	PurchasedWindowDays   float64
	ProfileWindowDays     float64
	ProfileBoost          float64
	ProfileTopNTags       int
	BlendAlpha            float64
	BlendBeta             float64
	BlendGamma            float64
	BanditAlgo            internaltypes.Algorithm
	Logger                *zap.Logger
	DecisionRecorder      audit.Recorder
	DecisionTraceSalt     string
	auditRecorderWarnOnce sync.Once
}

func (h *Handler) defaultOrgFromHeader(r *http.Request) uuid.UUID {
	return orgIDFromHeader(r, h.DefaultOrg)
}

func orgIDFromHeader(r *http.Request, fallback uuid.UUID) uuid.UUID {
	if r == nil {
		return fallback
	}
	org := r.Header.Get("X-Org-ID")
	if id, err := uuid.Parse(org); err == nil {
		return id
	}
	return fallback
}
