package store

import (
	"context"
	"testing"

	"recsys/internal/store"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestContentSimilarityTopK_ByTags(t *testing.T) {
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)

	s := store.New(pool)
	orgID := shared.MustOrgID(t)
	ns := "default"

	items := []store.ItemUpsert{
		{ItemID: "trail_pack", Available: true, Tags: []string{"trail", "running"}},
		{ItemID: "road_shoe", Available: true, Tags: []string{"road", "running"}},
		{ItemID: "gps_watch", Available: true, Tags: []string{"gps", "wearable"}},
	}
	require.NoError(t, s.UpsertItems(context.Background(), orgID, ns, items))

	// Request content candidates with overlapping tags.
	cands, err := s.ContentSimilarityTopK(
		context.Background(),
		orgID,
		ns,
		[]string{"running", "trail"},
		10,
		nil,
	)
	require.NoError(t, err)
	require.Len(t, cands, 2, "expected two items to match running/trail tags")
	require.Equal(t, "trail_pack", cands[0].ItemID, "item with two matching tags should rank first")
	require.Equal(t, 2.0, cands[0].Score, "score should reflect two overlapping tags")
	require.Equal(t, "road_shoe", cands[1].ItemID, "item with single overlapping tag should follow")
	require.Equal(t, 1.0, cands[1].Score)

	// Exclude the top result to validate filtering.
	excluded, err := s.ContentSimilarityTopK(
		context.Background(),
		orgID,
		ns,
		[]string{"running", "trail"},
		10,
		[]string{"trail_pack"},
	)
	require.NoError(t, err)
	require.Len(t, excluded, 1, "one candidate should remain after exclusion")
	require.Equal(t, "road_shoe", excluded[0].ItemID)
}
