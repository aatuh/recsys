package store

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aatuh/api-toolkit/v2/ports"
)

func TestTenantTablesMissingRLSReturnsMissingTables(t *testing.T) {
	t.Parallel()

	db := &fakeDBer{rows: &fakeRows{values: []string{"audit_log", "exposure_events"}}}
	got, err := TenantTablesMissingRLS(context.Background(), db, []string{"audit_log", "exposure_events"})
	if err != nil {
		t.Fatalf("TenantTablesMissingRLS() error = %v", err)
	}
	want := []string{"audit_log", "exposure_events"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TenantTablesMissingRLS() = %#v, want %#v", got, want)
	}
	if len(db.args) != 1 {
		t.Fatalf("query args len = %d", len(db.args))
	}
}

func TestTenantTablesMissingRLSHandlesNoRequiredTables(t *testing.T) {
	t.Parallel()

	got, err := TenantTablesMissingRLS(context.Background(), &fakeDBer{}, nil)
	if err != nil {
		t.Fatalf("TenantTablesMissingRLS() error = %v", err)
	}
	if got != nil {
		t.Fatalf("TenantTablesMissingRLS() = %#v, want nil", got)
	}
}

func TestTenantTablesMissingRLSPropagatesQueryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("db unavailable")
	_, err := TenantTablesMissingRLS(context.Background(), &fakeDBer{err: wantErr}, []string{"audit_log"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("TenantTablesMissingRLS() error = %v, want %v", err, wantErr)
	}
}

type fakeDBer struct {
	rows *fakeRows
	err  error
	args []any
}

func (f *fakeDBer) Exec(context.Context, string, ...any) (ports.DatabaseResult, error) {
	return nil, nil
}

func (f *fakeDBer) Query(_ context.Context, _ string, args ...any) (ports.DatabaseRows, error) {
	f.args = args
	if f.err != nil {
		return nil, f.err
	}
	if f.rows == nil {
		return &fakeRows{}, nil
	}
	return f.rows, nil
}

func (f *fakeDBer) QueryRow(context.Context, string, ...any) ports.DatabaseRow {
	return nil
}

type fakeRows struct {
	values []string
	idx    int
}

func (r *fakeRows) Next() bool {
	return r.idx < len(r.values)
}

func (r *fakeRows) Scan(dest ...any) error {
	*(dest[0].(*string)) = r.values[r.idx]
	r.idx++
	return nil
}

func (r *fakeRows) Close() {}

func (r *fakeRows) Err() error { return nil }
