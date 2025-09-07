package test

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
	s := store.New(pool)
	orgID := shared.MustOrgID(t)
	ns := "default"

	now := time.Now().UTC()

	// A should outrank B (more/stronger events for A)
	evs := []store.EventInsert{
		{UserID: "u1", ItemID: "A", Type: 0, Value: 1, TS: now},
		{UserID: "u1", ItemID: "A", Type: 3, Value: 1, TS: now}, // purchase weight 1.0
		{UserID: "u2", ItemID: "A", Type: 1, Value: 1, TS: now.Add(-1 * time.Hour)},
		{UserID: "u2", ItemID: "B", Type: 0, Value: 1, TS: now},
	}
	require.NoError(t, s.InsertEvents(context.Background(), orgID, ns, evs))

	got, err := s.PopularityTopK(context.Background(), orgID, ns, 30, 10)
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected popularity to return at least one item")
	require.Equal(t, "A", got[0].ItemID, "item A should rank above B given weights")
	if len(got) > 1 {
		require.Greater(t, got[0].Score, got[1].Score)
	}
}

func TestCooccurrenceTopK_Basics(t *testing.T) {
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)
	s := store.New(pool)
	orgID := shared.MustOrgID(t)
	ns := "default"

	now := time.Now().UTC()

	// u1 viewed B, then A, then C ; u2 viewed B and C
	evs := []store.EventInsert{
		{UserID: "u1", ItemID: "B", Type: 0, Value: 1, TS: now.Add(-2 * time.Hour)},
		{UserID: "u1", ItemID: "A", Type: 0, Value: 1, TS: now.Add(-90 * time.Minute)},
		{UserID: "u1", ItemID: "C", Type: 0, Value: 1, TS: now.Add(-30 * time.Minute)},
		{UserID: "u2", ItemID: "B", Type: 0, Value: 1, TS: now.Add(-70 * time.Minute)},
		{UserID: "u2", ItemID: "C", Type: 0, Value: 1, TS: now.Add(-10 * time.Minute)},
	}
	require.NoError(t, s.InsertEvents(context.Background(), orgID, ns, evs))

	got, err := s.CooccurrenceTopK(context.Background(), orgID, ns, "B", 10)
	require.NoError(t, err)
	require.NotEmpty(t, got, "co-vis for B should not be empty")
	ids := make([]string, 0, len(got))
	for _, g := range got {
		ids = append(ids, g.ItemID)
	}
	require.Contains(t, ids, "A")
	require.Contains(t, ids, "C")
}
