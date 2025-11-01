package handlers

import (
	"context"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/services/recommendation"
	"recsys/internal/types"
	handlerstypes "recsys/specs/types"

	"github.com/google/uuid"
)

type segmentStore interface {
	ListActiveSegmentsWithRules(ctx context.Context, orgID uuid.UUID, namespace string) ([]types.Segment, error)
	GetSegmentProfile(ctx context.Context, orgID uuid.UUID, namespace, profileID string) (*types.SegmentProfile, error)
	GetUser(ctx context.Context, orgID uuid.UUID, namespace, userID string) (*types.UserRecord, error)
}

func resolveSegmentSelection(
	ctx context.Context,
	store segmentStore,
	req algorithm.Request,
	httpReq handlerstypes.RecommendRequest,
	traitsOverride map[string]any,
) (recommendation.SegmentSelection, int64, error) {
	selection := recommendation.SegmentSelection{}

	segmentsList, err := store.ListActiveSegmentsWithRules(ctx, req.OrgID, req.Namespace)
	if err != nil {
		return selection, 0, err
	}
	if len(segmentsList) == 0 {
		return selection, 0, nil
	}

	traits := traitsOverride
	if traits == nil && req.UserID != "" {
		userRec, err := store.GetUser(ctx, req.OrgID, req.Namespace, req.UserID)
		if err != nil {
			return selection, 0, err
		}
		if userRec != nil && userRec.Traits != nil {
			traits = userRec.Traits
		}
	}

	data := buildSegmentContextData(req, httpReq.Context, traits)
	now := time.Now().UTC()

	var defaultSegment *types.Segment
	for i := range segmentsList {
		seg := segmentsList[i]
		if seg.SegmentID == "default" {
			defaultSegment = &segmentsList[i]
		}
		matchedRule, ok := segmentMatches(&seg, data, now)
		if !ok {
			continue
		}
		profile, err := store.GetSegmentProfile(ctx, req.OrgID, req.Namespace, seg.ProfileID)
		if err != nil {
			return selection, 0, err
		}
		selection.SegmentID = seg.SegmentID
		if profile != nil {
			selection.Profile = profile
			selection.ProfileID = profile.ProfileID
		}
		var ruleID int64
		if matchedRule != nil {
			ruleID = matchedRule.RuleID
		}
		return selection, ruleID, nil
	}

	if defaultSegment != nil {
		profile, err := store.GetSegmentProfile(ctx, req.OrgID, req.Namespace, defaultSegment.ProfileID)
		if err != nil {
			return selection, 0, err
		}
		selection.SegmentID = defaultSegment.SegmentID
		if profile != nil {
			selection.Profile = profile
			selection.ProfileID = profile.ProfileID
		}
		return selection, 0, nil
	}

	return selection, 0, nil
}
