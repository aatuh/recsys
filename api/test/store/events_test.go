//go:build integration

package store

import (
	"context"
	"recsys/shared/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := util.MustGetEnv("DATABASE_URL")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	return pool
}

func TestListUserEventsSince(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()
	s := New(pool)

	org := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ns := "test_purchased_" + time.Now().UTC().Format("20060102150405")
	uid := "u1"

	// Ensure items and users exist.
	price := 9.99
	items := []ItemUpsert{
		{ItemID: "i1", Available: true, Price: &price},
		{ItemID: "i2", Available: true, Price: &price},
		{ItemID: "i3", Available: true, Price: &price},
	}
	if err := s.UpsertItems(context.Background(), org, ns, items); err != nil {
		t.Fatalf("UpsertItems: %v", err)
	}
	if err := s.UpsertUsers(context.Background(), org, ns,
		[]UserUpsert{{UserID: uid}}); err != nil {
		t.Fatalf("UpsertUsers: %v", err)
	}

	// Insert purchases with different timestamps.
	now := time.Now().UTC()
	evs := []EventInsert{
		{UserID: uid, ItemID: "i1", Type: 3, Value: 1,
			TS: now.Add(-24 * time.Hour)},
		{UserID: uid, ItemID: "i2", Type: 3, Value: 1,
			TS: now.Add(-48 * time.Hour)},
		{UserID: uid, ItemID: "i3", Type: 0, Value: 1,
			TS: now.Add(-24 * time.Hour)}, // not a purchase
	}
	if err := s.InsertEvents(context.Background(), org, ns, evs); err != nil {
		t.Fatalf("InsertEvents: %v", err)
	}

	// Query since 36h -> expects i1 only when filtering to purchases.
	since := now.Add(-36 * time.Hour)
	got, err := s.ListUserEventsSince(context.Background(), org, ns, uid,
		since, []int16{3})
	if err != nil {
		t.Fatalf("ListUserEventsSince: %v", err)
	}
	if len(got) != 1 || got[0] != "i1" {
		t.Fatalf("want [i1]; got %v", got)
	}

	// Filtering to event type 0 should only return i3 within the window.
	got, err = s.ListUserEventsSince(context.Background(), org, ns, uid,
		since, []int16{0})
	if err != nil {
		t.Fatalf("ListUserEventsSince type 0: %v", err)
	}
	if len(got) != 1 || got[0] != "i3" {
		t.Fatalf("want [i3]; got %v", got)
	}

	// Nil event types should behave like "any" and include both i1 and i3.
	got, err = s.ListUserEventsSince(context.Background(), org, ns, uid,
		since, nil)
	if err != nil {
		t.Fatalf("ListUserEventsSince nil types: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 results; got %v", got)
	}
	seen := map[string]bool{}
	for _, id := range got {
		seen[id] = true
	}
	if !seen["i1"] || !seen["i3"] {
		t.Fatalf("want items i1 and i3; got %v", got)
	}
}
