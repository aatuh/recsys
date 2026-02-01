package validation

import (
	"testing"

	"github.com/aatuh/recsys-suite/api/src/specs/types"
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
