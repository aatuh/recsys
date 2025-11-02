package store

import (
	"context"
	"errors"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"recsys/internal/types"
)

//go:embed queries/manual_overrides_insert.sql
var manualOverrideInsertSQL string

//go:embed queries/manual_overrides_list.sql
var manualOverrideListSQL string

//go:embed queries/manual_overrides_cancel.sql
var manualOverrideCancelSQL string

//go:embed queries/manual_overrides_expire.sql
var manualOverrideExpireSQL string

func scanManualOverride(row pgx.Row) (types.ManualOverride, error) {
	var (
		record                        types.ManualOverride
		action                        string
		boostValue                    *float64
		ruleID                        *uuid.UUID
		status                        string
		notes, createdBy, cancelledBy *string
		expiresAt, cancelledAt        *time.Time
	)

	err := row.Scan(
		&record.OverrideID,
		&record.OrgID,
		&record.Namespace,
		&record.Surface,
		&action,
		&record.ItemID,
		&boostValue,
		&notes,
		&createdBy,
		&record.CreatedAt,
		&expiresAt,
		&ruleID,
		&status,
		&cancelledAt,
		&cancelledBy,
	)
	if err != nil {
		return types.ManualOverride{}, err
	}

	record.Action = types.ManualOverrideAction(action)
	if boostValue != nil {
		val := *boostValue
		record.BoostValue = &val
	}
	if notes != nil {
		record.Notes = *notes
	}
	if createdBy != nil {
		record.CreatedBy = *createdBy
	}
	if expiresAt != nil {
		at := expiresAt.UTC()
		record.ExpiresAt = &at
	}
	if ruleID != nil {
		rid := *ruleID
		record.RuleID = &rid
	}
	record.Status = types.ManualOverrideStatus(status)
	if cancelledAt != nil {
		at := cancelledAt.UTC()
		record.CancelledAt = &at
	}
	if cancelledBy != nil {
		record.CancelledBy = *cancelledBy
	}
	return record, nil
}

// CreateManualOverride inserts a manual override record.
func (s *Store) CreateManualOverride(ctx context.Context, override types.ManualOverride) (*types.ManualOverride, error) {
	var ruleID any
	if override.RuleID != nil {
		ruleID = *override.RuleID
	}
	row := s.Pool.QueryRow(ctx, manualOverrideInsertSQL,
		override.OverrideID,
		override.OrgID,
		override.Namespace,
		override.Surface,
		string(override.Action),
		override.ItemID,
		override.BoostValue,
		nullString(override.Notes),
		nullString(override.CreatedBy),
		nullTime(override.ExpiresAt),
		ruleID,
	)
	record, err := scanManualOverride(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListManualOverrides returns overrides filtered by namespace/surface/action/status.
func (s *Store) ListManualOverrides(
	ctx context.Context,
	orgID uuid.UUID,
	namespace string,
	surface string,
	filters types.ManualOverrideFilters,
) ([]types.ManualOverride, error) {
	if err := s.expireManualOverrides(ctx, orgID, time.Now().UTC()); err != nil {
		return nil, err
	}

	var (
		nsParam     any
		sfParam     any
		statusParam any
		actionParam any
	)
	if namespace != "" {
		nsParam = namespace
	}
	if surface != "" {
		sfParam = surface
	}
	if filters.Status != "" {
		statusParam = string(filters.Status)
	}
	if filters.Action != "" {
		actionParam = string(filters.Action)
	}

	rows, err := s.Pool.Query(ctx, manualOverrideListSQL,
		orgID,
		nsParam,
		sfParam,
		statusParam,
		actionParam,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []types.ManualOverride
	for rows.Next() {
		record, err := scanManualOverride(rows)
		if err != nil {
			return nil, err
		}
		if !filters.IncludeExpired && filters.Status != types.ManualOverrideStatusExpired && record.Status == types.ManualOverrideStatusExpired {
			continue
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// CancelManualOverride cancels an active override by ID.
func (s *Store) CancelManualOverride(ctx context.Context, orgID uuid.UUID, overrideID uuid.UUID, cancelledBy string) (*types.ManualOverride, error) {
	row := s.Pool.QueryRow(ctx, manualOverrideCancelSQL,
		overrideID,
		orgID,
		cancelledBy,
	)
	record, err := scanManualOverride(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (s *Store) expireManualOverrides(ctx context.Context, orgID uuid.UUID, now time.Time) error {
	_, err := s.Pool.Exec(ctx, manualOverrideExpireSQL, orgID, now)
	return err
}
