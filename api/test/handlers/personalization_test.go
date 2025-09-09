package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

// Test that Stage 5 personalization boosts items whose tags overlap the
// user's recent tag profile. We create a user with strong affinity to the
// "t:android" tag via recent purchases on Android-tagged items. Two
// candidate items (A,B) have equal base popularity, but only A matches
// "t:android", so A should rank above B and include a "personalization"
// reason in the response.
//
// Important: Profile-building items P1/P2 are set "available:false" so they
// cannot appear as recommendations themselves.
func TestRecommend_PersonalizationBoostsMatchingTags(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	// Items:
	// - A: candidate with t:android
	// - B: candidate with t:ios
	// - P1,P2: profile builders with t:android (unavailable for recs)
	do(http.MethodPost, "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "A", "available": true, "tags": []string{
				"category:phone", "t:android",
			}},
			{"item_id": "B", "available": true, "tags": []string{
				"category:phone", "t:ios",
			}},
			{"item_id": "P1", "available": false, "tags": []string{
				"category:accessory", "t:android",
			}},
			{"item_id": "P2", "available": false, "tags": []string{
				"category:accessory", "t:android",
			}},
		},
	}, http.StatusAccepted)

	// Users
	do(http.MethodPost, "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users":     []map[string]any{{"user_id": "u1"}, {"user_id": "u2"}},
	}, http.StatusAccepted)

	// Events:
	// - Equalize base popularity for A and B with simple view events so both
	//   are present as candidates without bias.
	// - Build u1's profile heavily toward t:android by purchasing P1 and P2.
	now := time.Now().UTC().Format(time.RFC3339)
	do(http.MethodPost, "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			// Candidate popularity (tie A vs B)
			{"user_id": "u2", "item_id": "A", "type": 0, "value": 1, "ts": now},
			{"user_id": "u2", "item_id": "B", "type": 0, "value": 1, "ts": now},

			// User profile (u1 loves android via purchases on P1/P2)
			{"user_id": "u1", "item_id": "P1", "type": 3, "value": 1, "ts": now},
			{"user_id": "u1", "item_id": "P2", "type": 3, "value": 1, "ts": now},
		},
	}, http.StatusAccepted)

	// Ask for recs with reasons. Expect A above B and A to include a
	// "personalization" reason.
	body := do(http.MethodPost, "/v1/recommendations", map[string]any{
		"user_id":         "u1",
		"namespace":       "default",
		"k":               2,
		"include_reasons": true,
	}, http.StatusOK)

	var resp struct {
		Items []struct {
			ItemID  string   `json:"item_id"`
			Score   float64  `json:"score"`
			Reasons []string `json:"reasons"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.GreaterOrEqual(t, len(resp.Items), 2)

	// A should be boosted above B due to personalization.
	require.Equal(t, "A", resp.Items[0].ItemID)

	// The top item should include a personalization reason.
	hasPersonalization := false
	for _, r := range resp.Items[0].Reasons {
		if strings.Contains(strings.ToLower(r), "personalization") {
			hasPersonalization = true
			break
		}
	}
	require.True(t, hasPersonalization,
		`expected "personalization" in reasons for top item`)
}

// Test that with no user history there is no personalization signal.
// We skew popularity toward B and request recs for a cold user; expect B
// on top and no "personalization" reason in any item.
func TestRecommend_Personalization_NoHistory_NoReason(t *testing.T) {
	client := shared.NewTestClient(t)

	// Clean database to ensure isolation
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	do := func(method, path string, body any, want int) []byte {
		return client.DoRequestWithStatus(t, method, path, body, want)
	}

	// Candidate items with distinct tags; B will be more popular.
	do(http.MethodPost, "/v1/items:upsert", map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{"item_id": "A", "available": true, "tags": []string{
				"category:phone", "t:android",
			}},
			{"item_id": "B", "available": true, "tags": []string{
				"category:phone", "t:ios",
			}},
		},
	}, http.StatusAccepted)

	// Users (u_new is cold; u_pop drives global popularity).
	do(http.MethodPost, "/v1/users:upsert", map[string]any{
		"namespace": "default",
		"users": []map[string]any{
			{"user_id": "u_new"},
			{"user_id": "u_pop"},
		},
	}, http.StatusAccepted)

	// Make B clearly more popular globally.
	now := time.Now().UTC().Format(time.RFC3339)
	do(http.MethodPost, "/v1/events:batch", map[string]any{
		"namespace": "default",
		"events": []map[string]any{
			{"user_id": "u_pop", "item_id": "B", "type": 3, "value": 1, "ts": now},
			{"user_id": "u_pop", "item_id": "B", "type": 1, "value": 1, "ts": now},
			{"user_id": "u_pop", "item_id": "A", "type": 0, "value": 1, "ts": now},
		},
	}, http.StatusAccepted)

	// Ask recs for cold user with reasons; expect B first and no
	// "personalization" reason present.
	body := do(http.MethodPost, "/v1/recommendations", map[string]any{
		"user_id":         "u_new",
		"namespace":       "default",
		"k":               2,
		"include_reasons": true,
	}, http.StatusOK)

	var resp struct {
		Items []struct {
			ItemID  string   `json:"item_id"`
			Score   float64  `json:"score"`
			Reasons []string `json:"reasons"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.GreaterOrEqual(t, len(resp.Items), 2)
	require.Equal(t, "B", resp.Items[0].ItemID)

	// None of the items should claim personalization for a cold user.
	for _, it := range resp.Items {
		for _, r := range it.Reasons {
			require.NotContains(t, strings.ToLower(r), "personalization",
				"unexpected personalization reason for cold user")
		}
	}
}
