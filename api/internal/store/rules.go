package store

import (
	"context"
	"errors"
	"time"

	_ "embed"

	"recsys/internal/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:embed queries/rules_insert.sql
var rulesInsertSQL string

//go:embed queries/rules_update.sql
var rulesUpdateSQL string

//go:embed queries/rules_get.sql
var rulesGetSQL string

//go:embed queries/rules_list.sql
var rulesListSQL string

//go:embed queries/rules_active_scope.sql
var rulesActiveScopeSQL string

type ruleScanner interface {
	Scan(dest ...any) error
}

func scanRule(scanner ruleScanner) (types.Rule, error) {
	var (
		rule       types.Rule
		desc       pgtype.Text
		action     string
		target     string
		targetKey  pgtype.Text
		itemIDs    []string
		boost      pgtype.Float8
		maxPins    pgtype.Int4
		segment    pgtype.Text
		validFrom  pgtype.Timestamptz
		validUntil pgtype.Timestamptz
	)

	if err := scanner.Scan(
		&rule.RuleID,
		&rule.OrgID,
		&rule.Namespace,
		&rule.Surface,
		&rule.Name,
		&desc,
		&action,
		&target,
		&targetKey,
		&itemIDs,
		&boost,
		&maxPins,
		&segment,
		&rule.Priority,
		&rule.Enabled,
		&validFrom,
		&validUntil,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		return types.Rule{}, err
	}

	if desc.Valid {
		rule.Description = desc.String
	}
	rule.Action = types.RuleAction(action)
	rule.TargetType = types.RuleTarget(target)
	if targetKey.Valid {
		rule.TargetKey = targetKey.String
	}
	if len(itemIDs) > 0 {
		rule.ItemIDs = append([]string(nil), itemIDs...)
	} else {
		rule.ItemIDs = []string{}
	}
	if boost.Valid {
		v := boost.Float64
		rule.BoostValue = &v
	}
	if maxPins.Valid {
		v := int(maxPins.Int32)
		rule.MaxPins = &v
	}
	if segment.Valid {
		rule.SegmentID = segment.String
	}
	if validFrom.Valid {
		t := validFrom.Time
		rule.ValidFrom = &t
	}
	if validUntil.Valid {
		t := validUntil.Time
		rule.ValidUntil = &t
	}
	return rule, nil
}

// CreateRule inserts a new rule and returns the stored representation.
func (s *Store) CreateRule(ctx context.Context, rule types.Rule) (*types.Rule, error) {
	row := s.Pool.QueryRow(ctx, rulesInsertSQL,
		rule.RuleID,
		rule.OrgID,
		rule.Namespace,
		rule.Surface,
		rule.Name,
		nullString(rule.Description),
		string(rule.Action),
		string(rule.TargetType),
		nullString(rule.TargetKey),
		rule.ItemIDs,
		nullFloat(rule.BoostValue),
		nullInt(rule.MaxPins),
		nullString(rule.SegmentID),
		rule.Priority,
		rule.Enabled,
		nullTime(rule.ValidFrom),
		nullTime(rule.ValidUntil),
	)
	stored, err := scanRule(row)
	if err != nil {
		return nil, err
	}
	return &stored, nil
}

// UpdateRule updates the rule identified by ruleID scoped to orgID.
func (s *Store) UpdateRule(ctx context.Context, orgID uuid.UUID, rule types.Rule) (*types.Rule, error) {
	row := s.Pool.QueryRow(ctx, rulesUpdateSQL,
		orgID,
		rule.Namespace,
		rule.Surface,
		rule.Name,
		nullString(rule.Description),
		string(rule.Action),
		string(rule.TargetType),
		nullString(rule.TargetKey),
		rule.ItemIDs,
		nullFloat(rule.BoostValue),
		nullInt(rule.MaxPins),
		nullString(rule.SegmentID),
		rule.Priority,
		rule.Enabled,
		nullTime(rule.ValidFrom),
		nullTime(rule.ValidUntil),
		rule.RuleID,
	)
	stored, err := scanRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &stored, nil
}

// GetRule retrieves a rule by ID scoped to orgID.
func (s *Store) GetRule(ctx context.Context, orgID uuid.UUID, ruleID uuid.UUID) (*types.Rule, error) {
	row := s.Pool.QueryRow(ctx, rulesGetSQL, orgID, ruleID)
	stored, err := scanRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &stored, nil
}

// ListRules returns rules filtered by optional parameters.
func (s *Store) ListRules(
	ctx context.Context,
	orgID uuid.UUID,
	namespace string,
	filters types.RuleListFilters,
) ([]types.Rule, error) {
	var (
		nsParam      any
		surfaceParam any
		segmentParam any
		enabledParam any
		activeParam  any
		actionParam  any
		targetParam  any
	)
	if namespace != "" {
		nsParam = namespace
	}
	if filters.Surface != "" {
		surfaceParam = filters.Surface
	}
	if filters.SegmentID != "" {
		segmentParam = filters.SegmentID
	}
	if filters.SegmentID == "__NULL__" {
		segmentParam = "__NULL__"
	}
	if filters.Enabled != nil {
		enabledParam = *filters.Enabled
	}
	if filters.ActiveAt != nil {
		activeParam = *filters.ActiveAt
	}
	if filters.Action != nil {
		actionParam = string(*filters.Action)
	}
	if filters.TargetType != nil {
		targetParam = string(*filters.TargetType)
	}

	rows, err := s.Pool.Query(ctx, rulesListSQL,
		orgID,
		nsParam,
		surfaceParam,
		segmentParam,
		enabledParam,
		activeParam,
		actionParam,
		targetParam,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []types.Rule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

// ListActiveRulesForScope returns enabled rules for a namespace/surface/segment combination.
func (s *Store) ListActiveRulesForScope(
	ctx context.Context,
	orgID uuid.UUID,
	namespace, surface, segmentID string,
	ts time.Time,
) ([]types.Rule, error) {
	tsProvided := !ts.IsZero()
	if !tsProvided {
		ts = time.Now().UTC()
	}
	rows, err := s.Pool.Query(ctx, rulesActiveScopeSQL,
		orgID,
		namespace,
		surface,
		nullString(segmentID),
		tsProvided,
		ts,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []types.Rule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullFloat(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullTime(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
