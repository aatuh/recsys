package store

import (
	"context"
	"testing"
	"time"

	"recsys/internal/http/store"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestPopularityTopK_Basics(t *testing.T) {
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	shared.MustHaveEventTypeDefaults(t, pool)

	s := store.New(pool)
	orgID := shared.MustOrgID(t)
	ns := "default"
	now := time.Now().UTC()

	// Ensure items exist and are available; popularity query joins items with
	// i.available=true.
	require.NoError(t, s.UpsertItems(
		context.Background(), orgID, ns,
		[]store.ItemUpsert{
			{ItemID: "A", Available: true},
			{ItemID: "B", Available: true},
		},
	))

	// A should outrank B due to stronger/more events and default weights:
	// view=0.1, click=0.3, purchase=1.0.
	evs := []store.EventInsert{
		{UserID: "u1", ItemID: "A", Type: 0, Value: 1, TS: now},
		{UserID: "u1", ItemID: "A", Type: 3, Value: 1, TS: now},
		{UserID: "u2", ItemID: "A", Type: 1, Value: 1, TS: now.Add(-1 * time.Hour)},
		{UserID: "u2", ItemID: "B", Type: 0, Value: 1, TS: now},
	}
	require.NoError(t, s.InsertEvents(context.Background(), orgID, ns, evs))

	got, err := s.PopularityTopK(context.Background(), orgID, ns, 30, 10, nil)
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected popularity to return at least one item")
	require.Equal(t, "A", got[0].ItemID, "item A should rank above B given weights")
	if len(got) > 1 {
		require.Greater(t, got[0].Score, got[1].Score)
	}
}
