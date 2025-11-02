package manual

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"recsys/internal/types"
)

// Store defines persistence operations required by the manual override service.
type Store interface {
	CreateManualOverride(ctx context.Context, override types.ManualOverride) (*types.ManualOverride, error)
	ListManualOverrides(ctx context.Context, orgID uuid.UUID, namespace string, surface string, filters types.ManualOverrideFilters) ([]types.ManualOverride, error)
	CancelManualOverride(ctx context.Context, orgID uuid.UUID, overrideID uuid.UUID, cancelledBy string) (*types.ManualOverride, error)
	CreateRule(ctx context.Context, rule types.Rule) (*types.Rule, error)
	UpdateRule(ctx context.Context, orgID uuid.UUID, rule types.Rule) (*types.Rule, error)
	GetRule(ctx context.Context, orgID uuid.UUID, ruleID uuid.UUID) (*types.Rule, error)
}

// Service orchestrates manual merchandising overrides.
type Service struct {
	store Store
	now   func() time.Time
}

// Option configures the service.
type Option func(*Service)

// WithNow overrides the time source (useful for tests).
func WithNow(now func() time.Time) Option {
	return func(s *Service) {
		if now != nil {
			s.now = now
		}
	}
}

// New constructs a manual overrides service.
func New(store Store, opts ...Option) *Service {
	svc := &Service{
		store: store,
		now:   time.Now,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// CreateInput captures the request to create a manual override.
type CreateInput struct {
	Namespace  string
	Surface    string
	ItemID     string
	Action     types.ManualOverrideAction
	BoostValue *float64
	Notes      string
	CreatedBy  string
	ExpiresAt  *time.Time
	Priority   *int
}

const defaultManualPriority = 1000
const defaultManualBoost = 1.0

// Create registers a manual override and corresponding rule.
func (s *Service) Create(ctx context.Context, orgID uuid.UUID, input CreateInput) (*types.ManualOverride, error) {
	if s.store == nil {
		return nil, errors.New("manual override service: store is nil")
	}
	if strings.TrimSpace(input.Namespace) == "" {
		return nil, errors.New("namespace is required")
	}
	if strings.TrimSpace(input.ItemID) == "" {
		return nil, errors.New("item_id is required")
	}
	if strings.TrimSpace(input.Surface) == "" {
		input.Surface = "default"
	}
	if input.Action != types.ManualOverrideActionBoost && input.Action != types.ManualOverrideActionSuppress {
		return nil, fmt.Errorf("unsupported action %q", input.Action)
	}
	if input.Action == types.ManualOverrideActionBoost {
		if input.BoostValue != nil && *input.BoostValue <= 0 {
			return nil, errors.New("boost_value must be positive for boost overrides")
		}
		if input.BoostValue == nil {
			val := defaultManualBoost
			input.BoostValue = &val
		}
	} else if input.BoostValue != nil {
		return nil, errors.New("boost_value is only valid for boost overrides")
	}

	now := s.now().UTC()
	priority := defaultManualPriority
	if input.Priority != nil && *input.Priority > 0 {
		priority = *input.Priority
	}

	var validUntil *time.Time
	if input.ExpiresAt != nil {
		exp := input.ExpiresAt.UTC()
		validUntil = &exp
	}

	ruleID := uuid.New()
	rule := types.Rule{
		RuleID:      ruleID,
		OrgID:       orgID,
		Namespace:   strings.TrimSpace(input.Namespace),
		Surface:     strings.TrimSpace(input.Surface),
		Name:        fmt.Sprintf("manual_%s_%s", string(input.Action), input.ItemID),
		Description: strings.TrimSpace(input.Notes),
		TargetType:  types.RuleTargetItem,
		TargetKey:   input.ItemID,
		ItemIDs:     []string{input.ItemID},
		Priority:    priority,
		Enabled:     true,
		ValidFrom:   &now,
		ValidUntil:  validUntil,
	}

	switch input.Action {
	case types.ManualOverrideActionBoost:
		rule.Action = types.RuleActionBoost
		rule.BoostValue = input.BoostValue
	case types.ManualOverrideActionSuppress:
		rule.Action = types.RuleActionBlock
	default:
		return nil, fmt.Errorf("unsupported action %q", input.Action)
	}

	createdRule, err := s.store.CreateRule(ctx, rule)
	if err != nil {
		return nil, err
	}

	override := types.ManualOverride{
		OverrideID: uuid.New(),
		OrgID:      orgID,
		Namespace:  rule.Namespace,
		Surface:    rule.Surface,
		ItemID:     input.ItemID,
		Action:     input.Action,
		BoostValue: input.BoostValue,
		Notes:      strings.TrimSpace(input.Notes),
		CreatedBy:  strings.TrimSpace(input.CreatedBy),
		ExpiresAt:  validUntil,
		RuleID:     &createdRule.RuleID,
	}

	record, err := s.store.CreateManualOverride(ctx, override)
	if err != nil {
		// best-effort: disable the rule if we failed to persist the override
		if createdRule != nil {
			createdRule.Enabled = false
			_, _ = s.store.UpdateRule(ctx, orgID, *createdRule)
		}
		return nil, err
	}
	return record, nil
}

// List returns manual overrides for the namespace/surface.
func (s *Service) List(ctx context.Context, orgID uuid.UUID, namespace, surface string, filters types.ManualOverrideFilters) ([]types.ManualOverride, error) {
	if s.store == nil {
		return nil, errors.New("manual override service: store is nil")
	}
	return s.store.ListManualOverrides(ctx, orgID, namespace, surface, filters)
}

// Cancel disables an active manual override and its backing rule.
func (s *Service) Cancel(ctx context.Context, orgID uuid.UUID, overrideID uuid.UUID, cancelledBy string) (*types.ManualOverride, error) {
	if s.store == nil {
		return nil, errors.New("manual override service: store is nil")
	}
	record, err := s.store.CancelManualOverride(ctx, orgID, overrideID, cancelledBy)
	if err != nil || record == nil {
		return record, err
	}
	if record.RuleID != nil {
		existing, err := s.store.GetRule(ctx, orgID, *record.RuleID)
		if err == nil && existing != nil {
			existing.Enabled = false
			_, _ = s.store.UpdateRule(ctx, orgID, *existing)
		}
	}
	return record, nil
}
