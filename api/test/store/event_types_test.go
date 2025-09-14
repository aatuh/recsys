package store

import (
	"context"
	"testing"
	"time"

	"recsys/internal/store"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestPopularity_OverridesFlipRanking(t *testing.T) {
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	s := store.New(pool)
	org := shared.MustOrgID(t)
	ns := "default"

	// Ensure items exist, because PopularityTopK joins against items and
	// requires availability=true.
	require.NoError(t, s.UpsertItems(
		context.Background(),
		org,
		ns,
		[]store.ItemUpsert{
			{ItemID: "A", Available: true},
			{ItemID: "B", Available: true},
		},
	))

	now := time.Now().UTC()
	evs := []store.EventInsert{
		{UserID: "u1", ItemID: "A", Type: 0, Value: 1, TS: now}, // view (0.1)
		{UserID: "u1", ItemID: "B", Type: 3, Value: 1, TS: now}, // purchase (1.0)
	}
	require.NoError(t, s.InsertEvents(context.Background(), org, ns, evs))

	top, err := s.PopularityTopK(context.Background(), org, ns, 14, 2, nil)
	require.NoError(t, err)
	require.NotEmpty(t, top)
	require.Equal(t, "B", top[0].ItemID, "purchase should outrank view by default")

	// Boost view, reduce purchase, then the ranking should flip.
	require.NoError(t, s.UpsertEventTypeConfig(
		context.Background(),
		org,
		ns,
		[]store.EventTypeConfig{
			{Type: 0, Weight: 1.2},
			{Type: 3, Weight: 0.4},
		},
	))

	top2, err := s.PopularityTopK(context.Background(), org, ns, 14, 2, nil)
	require.NoError(t, err)
	require.NotEmpty(t, top2)
	require.Equal(t, "A", top2[0].ItemID, "override should flip ranking")
}
