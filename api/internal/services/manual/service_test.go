package manual

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"recsys/internal/types"
)

type stubStore struct {
	createdRule       *types.Rule
	createdOverride   *types.ManualOverride
	updatedRule       *types.Rule
	listResult        []types.ManualOverride
	getRule           *types.Rule
	createRuleErr     error
	createOverrideErr error
	cancelResult      *types.ManualOverride
}

func (s *stubStore) CreateManualOverride(ctx context.Context, override types.ManualOverride) (*types.ManualOverride, error) {
	if s.createOverrideErr != nil {
		return nil, s.createOverrideErr
	}
	dup := override
	dup.CreatedAt = time.Now().UTC()
	dup.Status = types.ManualOverrideStatusActive
	s.createdOverride = &dup
	return &dup, nil
}

func (s *stubStore) ListManualOverrides(ctx context.Context, orgID uuid.UUID, namespace string, surface string, filters types.ManualOverrideFilters) ([]types.ManualOverride, error) {
	return s.listResult, nil
}

func (s *stubStore) CancelManualOverride(ctx context.Context, orgID uuid.UUID, overrideID uuid.UUID, cancelledBy string) (*types.ManualOverride, error) {
	return s.cancelResult, nil
}

func (s *stubStore) CreateRule(ctx context.Context, rule types.Rule) (*types.Rule, error) {
	if s.createRuleErr != nil {
		return nil, s.createRuleErr
	}
	s.createdRule = &rule
	return s.createdRule, nil
}

func (s *stubStore) UpdateRule(ctx context.Context, orgID uuid.UUID, rule types.Rule) (*types.Rule, error) {
	s.updatedRule = &rule
	return &rule, nil
}

func (s *stubStore) GetRule(ctx context.Context, orgID uuid.UUID, ruleID uuid.UUID) (*types.Rule, error) {
	return s.getRule, nil
}

func TestServiceCreateBoost(t *testing.T) {
	store := &stubStore{}
	svc := New(store)
	orgID := uuid.New()
	boost := 2.0
	input := CreateInput{
		Namespace:  "default",
		Surface:    "home",
		ItemID:     "sku-1",
		Action:     types.ManualOverrideActionBoost,
		BoostValue: &boost,
		Notes:      "seasonal push",
		CreatedBy:  "buyer@example.com",
	}

	record, err := svc.Create(context.Background(), orgID, input)
	require.NoError(t, err)
	require.NotNil(t, record)
	require.NotNil(t, store.createdRule)
	require.Equal(t, types.RuleActionBoost, store.createdRule.Action)
	require.NotNil(t, store.createdRule.BoostValue)
	require.Equal(t, boost, *store.createdRule.BoostValue)
	require.Equal(t, record.RuleID.String(), store.createdRule.RuleID.String())
}

func TestServiceCancelDisablesRule(t *testing.T) {
	ruleID := uuid.New()
	store := &stubStore{
		cancelResult: &types.ManualOverride{
			OverrideID: uuid.New(),
			OrgID:      uuid.New(),
			RuleID:     &ruleID,
		},
		getRule: &types.Rule{
			RuleID:    ruleID,
			Enabled:   true,
			Namespace: "default",
			Surface:   "home",
		},
	}
	svc := New(store)

	record, err := svc.Cancel(context.Background(), store.cancelResult.OrgID, store.cancelResult.OverrideID, "ops@example.com")
	require.NoError(t, err)
	require.NotNil(t, record)
	require.NotNil(t, store.updatedRule)
	require.False(t, store.updatedRule.Enabled)
}

func TestServiceCreateValidation(t *testing.T) {
	svc := New(&stubStore{})
	_, err := svc.Create(context.Background(), uuid.New(), CreateInput{})
	require.Error(t, err)
	_, err = svc.Create(context.Background(), uuid.New(), CreateInput{Namespace: "default", ItemID: "sku", Action: "invalid"})
	require.Error(t, err)
	_, err = svc.Create(context.Background(), uuid.New(), CreateInput{Namespace: "default", ItemID: "sku", Action: types.ManualOverrideActionBoost, BoostValue: func() *float64 { v := -1.0; return &v }()})
	require.Error(t, err)
}

func TestServiceCreateHandlesRuleFailure(t *testing.T) {
	store := &stubStore{createRuleErr: errors.New("db down")}
	svc := New(store)
	_, err := svc.Create(context.Background(), uuid.New(), CreateInput{Namespace: "default", ItemID: "sku", Action: types.ManualOverrideActionSuppress})
	require.Error(t, err)
}
