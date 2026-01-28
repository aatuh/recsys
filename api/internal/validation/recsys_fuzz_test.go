package validation

import (
	"testing"

	"recsys/src/specs/types"
)

func FuzzNormalizeRecommendRequest(f *testing.F) {
	f.Add("home", "u_1")
	f.Fuzz(func(t *testing.T, surface, userID string) {
		req := &types.RecommendRequest{
			Surface: surface,
			User: &types.UserRef{
				UserID: userID,
			},
		}
		_, _, _ = NormalizeRecommendRequest(req)
	})
}
