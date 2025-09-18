package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	_ "embed"

	"recsys/internal/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:embed queries/segment_profiles_list.sql
var segmentProfilesListSQL string

//go:embed queries/segment_profile_get.sql
var segmentProfileGetSQL string

//go:embed queries/segment_profiles_upsert.sql
var segmentProfilesUpsertSQL string

//go:embed queries/segment_profiles_delete.sql
var segmentProfilesDeleteSQL string

//go:embed queries/segments_list_with_rules.sql
var segmentsListWithRulesSQL string

//go:embed queries/segments_active_with_rules.sql
var segmentsActiveWithRulesSQL string

//go:embed queries/segments_upsert.sql
var segmentsUpsertSQL string

//go:embed queries/segments_delete.sql
var segmentsDeleteSQL string

//go:embed queries/segment_rules_insert.sql
var segmentRulesInsertSQL string

//go:embed queries/segment_rules_delete_by_segment.sql
var segmentRulesDeleteBySegmentSQL string

// ListSegmentProfiles returns all profiles for the namespace.
func (s *Store) ListSegmentProfiles(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
) ([]types.SegmentProfile, error) {
	rows, err := s.Pool.Query(ctx, segmentProfilesListSQL, orgID, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []types.SegmentProfile
	for rows.Next() {
		var profile types.SegmentProfile
		var excludeEventTypes []int16
		var brandPrefixes []string
		var categoryPrefixes []string
		if err := rows.Scan(
			&profile.ProfileID,
			&profile.Description,
			&profile.BlendAlpha,
			&profile.BlendBeta,
			&profile.BlendGamma,
			&profile.MMRLambda,
			&profile.BrandCap,
			&profile.CategoryCap,
			&profile.ProfileBoost,
			&profile.ProfileWindowDays,
			&profile.ProfileTopN,
			&profile.HalfLifeDays,
			&profile.CoVisWindowDays,
			&profile.PurchasedWindowDays,
			&profile.RuleExcludeEvents,
			&excludeEventTypes,
			&brandPrefixes,
			&categoryPrefixes,
			&profile.PopularityFanout,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profile.ExcludeEventTypes = append([]int16(nil), excludeEventTypes...)
		profile.BrandTagPrefixes = append([]string(nil), brandPrefixes...)
		profile.CategoryTagPrefixes = append([]string(nil), categoryPrefixes...)
		out = append(out, profile)
	}

	return out, rows.Err()
}

// GetSegmentProfile returns a profile by ID.
func (s *Store) GetSegmentProfile(
	ctx context.Context,
	orgID uuid.UUID,
	ns, profileID string,
) (*types.SegmentProfile, error) {
	row := s.Pool.QueryRow(ctx, segmentProfileGetSQL, orgID, ns, profileID)
	var profile types.SegmentProfile
	var excludeEventTypes []int16
	var brandPrefixes []string
	var categoryPrefixes []string
	if err := row.Scan(
		&profile.ProfileID,
		&profile.Description,
		&profile.BlendAlpha,
		&profile.BlendBeta,
		&profile.BlendGamma,
		&profile.MMRLambda,
		&profile.BrandCap,
		&profile.CategoryCap,
		&profile.ProfileBoost,
		&profile.ProfileWindowDays,
		&profile.ProfileTopN,
		&profile.HalfLifeDays,
		&profile.CoVisWindowDays,
		&profile.PurchasedWindowDays,
		&profile.RuleExcludeEvents,
		&excludeEventTypes,
		&brandPrefixes,
		&categoryPrefixes,
		&profile.PopularityFanout,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	profile.ExcludeEventTypes = append([]int16(nil), excludeEventTypes...)
	profile.BrandTagPrefixes = append([]string(nil), brandPrefixes...)
	profile.CategoryTagPrefixes = append([]string(nil), categoryPrefixes...)
	return &profile, nil
}

// UpsertSegmentProfiles stores or updates the provided profiles.
func (s *Store) UpsertSegmentProfiles(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	profiles []types.SegmentProfile,
) error {
	if len(profiles) == 0 {
		return nil
	}
	bat := &pgx.Batch{}
	for _, profile := range profiles {
		bat.Queue(
			segmentProfilesUpsertSQL,
			orgID,
			ns,
			profile.ProfileID,
			profile.Description,
			profile.BlendAlpha,
			profile.BlendBeta,
			profile.BlendGamma,
			profile.MMRLambda,
			profile.BrandCap,
			profile.CategoryCap,
			profile.ProfileBoost,
			profile.ProfileWindowDays,
			profile.ProfileTopN,
			profile.HalfLifeDays,
			profile.CoVisWindowDays,
			profile.PurchasedWindowDays,
			profile.RuleExcludeEvents,
			profile.ExcludeEventTypes,
			profile.BrandTagPrefixes,
			profile.CategoryTagPrefixes,
			profile.PopularityFanout,
		)
	}
	br := s.Pool.SendBatch(ctx, bat)
	defer br.Close()
	for range profiles {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// DeleteSegmentProfiles removes profiles by ID.
func (s *Store) DeleteSegmentProfiles(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	profileIDs []string,
) (int64, error) {
	if len(profileIDs) == 0 {
		return 0, nil
	}
	cmd, err := s.Pool.Exec(ctx, segmentProfilesDeleteSQL, orgID, ns, profileIDs)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

// ListSegmentsWithRules lists all segments including their rules.
func (s *Store) ListSegmentsWithRules(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
) ([]types.Segment, error) {
	rows, err := s.Pool.Query(ctx, segmentsListWithRulesSQL, orgID, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := map[string]*types.Segment{}
	order := []string{}
	for rows.Next() {
		var (
			segmentID   string
			name        string
			priority    int
			active      bool
			profileID   string
			description string
			createdAt   time.Time
			updatedAt   time.Time
			ruleID      pgtype.Int8
			ruleJSON    []byte
			ruleEnabled pgtype.Bool
			ruleDesc    pgtype.Text
			ruleCreated pgtype.Timestamptz
			ruleUpdated pgtype.Timestamptz
		)

		if err := rows.Scan(
			&segmentID,
			&name,
			&priority,
			&active,
			&profileID,
			&description,
			&createdAt,
			&updatedAt,
			&ruleID,
			&ruleJSON,
			&ruleEnabled,
			&ruleDesc,
			&ruleCreated,
			&ruleUpdated,
		); err != nil {
			return nil, err
		}

		seg, ok := segments[segmentID]
		if !ok {
			seg = &types.Segment{
				SegmentID:   segmentID,
				Name:        name,
				Priority:    priority,
				Active:      active,
				ProfileID:   profileID,
				Description: description,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}
			segments[segmentID] = seg
			order = append(order, segmentID)
		}

		if ruleID.Valid && ruleJSON != nil {
			ruleBytes := append([]byte(nil), ruleJSON...)
			desc := ""
			if ruleDesc.Valid {
				desc = ruleDesc.String
			}
			rule := types.SegmentRule{
				RuleID:      ruleID.Int64,
				Rule:        json.RawMessage(ruleBytes),
				Enabled:     ruleEnabled.Valid && ruleEnabled.Bool,
				Description: desc,
				CreatedAt:   valueOrDefaultTime(ruleCreated),
				UpdatedAt:   valueOrDefaultTime(ruleUpdated),
			}
			seg.Rules = append(seg.Rules, rule)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []types.Segment
	for _, id := range order {
		out = append(out, *segments[id])
	}
	return out, nil
}

// ListActiveSegmentsWithRules returns only active segments and enabled rules.
func (s *Store) ListActiveSegmentsWithRules(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
) ([]types.Segment, error) {
	rows, err := s.Pool.Query(ctx, segmentsActiveWithRulesSQL, orgID, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := map[string]*types.Segment{}
	order := []string{}
	for rows.Next() {
		var (
			segmentID   string
			name        string
			priority    int
			profileID   string
			description string
			ruleID      pgtype.Int8
			ruleJSON    []byte
			ruleEnabled pgtype.Bool
		)

		if err := rows.Scan(
			&segmentID,
			&name,
			&priority,
			&profileID,
			&description,
			&ruleID,
			&ruleJSON,
			&ruleEnabled,
		); err != nil {
			return nil, err
		}

		seg, ok := segments[segmentID]
		if !ok {
			seg = &types.Segment{
				SegmentID:   segmentID,
				Name:        name,
				Priority:    priority,
				Active:      true,
				ProfileID:   profileID,
				Description: description,
			}
			segments[segmentID] = seg
			order = append(order, segmentID)
		}

		if ruleID.Valid && ruleJSON != nil && ruleEnabled.Valid && ruleEnabled.Bool {
			ruleBytes := append([]byte(nil), ruleJSON...)
			seg.Rules = append(seg.Rules, types.SegmentRule{
				RuleID:  ruleID.Int64,
				Rule:    json.RawMessage(ruleBytes),
				Enabled: true,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []types.Segment
	for _, id := range order {
		out = append(out, *segments[id])
	}
	return out, nil
}

// UpsertSegmentWithRules upserts the segment and replaces its rules atomically.
func (s *Store) UpsertSegmentWithRules(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	segment types.Segment,
) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, segmentsUpsertSQL,
		orgID, ns, segment.SegmentID, segment.Name, segment.Priority,
		segment.Active, segment.ProfileID, segment.Description,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, segmentRulesDeleteBySegmentSQL, orgID, ns, segment.SegmentID); err != nil {
		return err
	}

	for _, rule := range segment.Rules {
		if len(rule.Rule) == 0 {
			continue
		}
		if _, err := tx.Exec(ctx, segmentRulesInsertSQL,
			orgID,
			ns,
			segment.SegmentID,
			[]byte(rule.Rule),
			rule.Enabled,
			rule.Description,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// DeleteSegments removes the provided segments (cascading rules).
func (s *Store) DeleteSegments(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	segmentIDs []string,
) (int64, error) {
	if len(segmentIDs) == 0 {
		return 0, nil
	}
	cmd, err := s.Pool.Exec(ctx, segmentsDeleteSQL, orgID, ns, segmentIDs)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func valueOrDefaultTime(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}
