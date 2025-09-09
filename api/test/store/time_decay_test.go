package store

import (
	"context"
	"testing"
	"time"

	"recsys/internal/http/store"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestPopularity_TimeDecayMakesRecentWin(t *testing.T) {
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	s := store.New(pool)
	org := shared.MustOrgID(t)
	ns := "default"

	now := time.Now().UTC()

	// Ensure items exist; popularity query joins items with available=true.
	require.NoError(t, s.UpsertItems(
		context.Background(),
		org,
		ns,
		[]store.ItemUpsert{
			{ItemID: "A", Available: true},
			{ItemID: "B", Available: true},
		},
	))

	// A is old; B is recent. With short half-life, B should win.
	evs := []store.EventInsert{
		{
			UserID: "u1", ItemID: "A", Type: 0, Value: 10,
			TS: now.Add(-14 * 24 * time.Hour),
		},
		{
			UserID: "u1", ItemID: "B", Type: 0, Value: 1,
			TS: now,
		},
	}
	require.NoError(t, s.InsertEvents(context.Background(), org, ns, evs))

	top, err := s.PopularityTopK(context.Background(), org, ns, 3, 2, nil)
	require.NoError(t, err)
	require.NotEmpty(t, top)
	require.Equal(t, "B", top[0].ItemID,
		"recent item should outrank older despite magnitude due to decay")
}
