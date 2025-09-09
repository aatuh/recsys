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

func newTestPool2(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
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

func TestCooccurrenceTopKWithin_Basics(t *testing.T) {
	pool := newTestPool2(t)
	defer pool.Close()
	s := New(pool)

	org := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ns := "test_covis_" + time.Now().UTC().Format("20060102150405")

	// Items and users.
	price := 1.0
	items := []ItemUpsert{
		{ItemID: "A", Available: true, Price: &price},
		{ItemID: "B", Available: true, Price: &price},
		{ItemID: "C", Available: true, Price: &price},
	}
	if err := s.UpsertItems(context.Background(), org, ns, items); err != nil {
		t.Fatalf("UpsertItems: %v", err)
	}
	if err := s.UpsertUsers(context.Background(), org, ns,
		[]UserUpsert{{UserID: "u1"}, {UserID: "u2"}}); err != nil {
		t.Fatalf("UpsertUsers: %v", err)
	}

	now := time.Now().UTC()
	// u1 views A, then B and C
	// u2 views A, then B
	evs := []EventInsert{
		{UserID: "u1", ItemID: "A", Type: 0, TS: now.Add(-1 * time.Hour)},
		{UserID: "u1", ItemID: "B", Type: 0, TS: now.Add(-50 * time.Minute)},
		{UserID: "u1", ItemID: "C", Type: 0, TS: now.Add(-40 * time.Minute)},
		{UserID: "u2", ItemID: "A", Type: 0, TS: now.Add(-30 * time.Minute)},
		{UserID: "u2", ItemID: "B", Type: 0, TS: now.Add(-20 * time.Minute)},
	}
	if err := s.InsertEvents(context.Background(), org, ns, evs); err != nil {
		t.Fatalf("InsertEvents: %v", err)
	}

	// since = 2h ago, so all events count.
	since := now.Add(-2 * time.Hour)
	got, err := s.CooccurrenceTopKWithin(context.Background(), org, ns, "A",
		5, since)
	if err != nil {
		t.Fatalf("CooccurrenceTopKWithin: %v", err)
	}

	// B should have higher count than C (2 vs 1).
	if len(got) < 2 || got[0].ItemID != "B" || got[0].Score <= got[1].Score {
		t.Fatalf("unexpected order: got=%v", got)
	}
}
