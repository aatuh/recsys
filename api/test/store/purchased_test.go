//go:build integration

package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestDBURL() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return os.Getenv("DATABASE_URL")
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := getTestDBURL()
	if url == "" {
		t.Skip("no TEST_DATABASE_URL/DATABASE_URL; skipping")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	return pool
}

func TestListUserPurchasedSince(t *testing.T) {
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

	// Query since 36h -> expects i1 only.
	since := now.Add(-36 * time.Hour)
	got, err := s.ListUserPurchasedSince(context.Background(), org, ns, uid,
		since)
	if err != nil {
		t.Fatalf("ListUserPurchasedSince: %v", err)
	}
	if len(got) != 1 || got[0] != "i1" {
		t.Fatalf("want [i1]; got %v", got)
	}
}
