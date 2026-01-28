package store

import (
	"context"

	"recsys/internal/services/foosvc"

	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/ports"
)

// FooRepo is a Postgres adapter for foosvc.Repo.
type FooRepo struct {
	Pool ports.DatabasePool
}

func NewFooRepo(pool ports.DatabasePool) *FooRepo {
	return &FooRepo{Pool: pool}
}

func (r *FooRepo) Create(ctx context.Context, f *foosvc.Foo) error {
	db := txpostgres.FromCtx(ctx, r.Pool)
	const q = `
	insert into foo (id, org_id, namespace, name, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(ctx, q,
		f.ID, f.OrgID, f.Namespace, f.Name, f.CreatedAt, f.UpdatedAt)
	if isUniqueViolation(err) {
		return foosvc.ErrConflict
	}
	return err
}

func (r *FooRepo) GetByID(
	ctx context.Context, id string,
) (*foosvc.Foo, error) {
	db := txpostgres.FromCtx(ctx, r.Pool)
	const q = `
	select id, org_id, namespace, name, created_at, updated_at
	from foo where id=$1
	`
	var f foosvc.Foo
	err := db.QueryRow(ctx, q, id).Scan(
		&f.ID, &f.OrgID, &f.Namespace, &f.Name, &f.CreatedAt, &f.UpdatedAt,
	)
	if txpostgres.IsNoRows(err) {
		return nil, foosvc.ErrNotFound
	}
	return &f, err
}

func (r *FooRepo) Update(ctx context.Context, f *foosvc.Foo) error {
	db := txpostgres.FromCtx(ctx, r.Pool)
	const q = `
	update foo
	set name=$2, updated_at=$3
	where id=$1
	`
	ct, err := db.Exec(ctx, q, f.ID, f.Name, f.UpdatedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return foosvc.ErrNotFound
	}
	return nil
}

func (r *FooRepo) Delete(ctx context.Context, id string) error {
	db := txpostgres.FromCtx(ctx, r.Pool)
	const q = `delete from foo where id=$1`
	ct, err := db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return foosvc.ErrNotFound
	}
	return nil
}

func (r *FooRepo) List(
	ctx context.Context, orgID, ns string,
	limit, offset int, search string,
) ([]foosvc.Foo, int, error) {
	db := txpostgres.FromCtx(ctx, r.Pool)
	const q = `
	select id, org_id, namespace, name, created_at, updated_at,
	       count(*) over() as total_count
	from foo
	where org_id=$1 and namespace=$2 and
	      ($3='' or name ilike '%'||$3||'%')
	order by created_at desc
	limit $4 offset $5
	`
	rows, err := db.Query(ctx, q, orgID, ns, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []foosvc.Foo
	var total int
	for rows.Next() {
		var f foosvc.Foo
		var totalCount int64
		if err := rows.Scan(
			&f.ID, &f.OrgID, &f.Namespace, &f.Name,
			&f.CreatedAt, &f.UpdatedAt, &totalCount,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, f)
		total = int(totalCount)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// Generic fallback; pg-specific detection lives in toolkit if needed.
	return containsAny(err.Error(), "unique", "duplicate", "23505")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && stringContainsFold(s, sub) {
			return true
		}
	}
	return false
}

func stringContainsFold(s, sub string) bool {
	// Minimal allocation-less folded contains.
	ls, lsub := len(s), len(sub)
	if lsub == 0 || lsub > ls {
		return false
	}
	for i := 0; i <= ls-lsub; i++ {
		match := true
		for j := 0; j < lsub; j++ {
			a := s[i+j]
			b := sub[j]
			// fold ascii only
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Ensure FooRepo satisfies the domain port.
var _ foosvc.Repo = (*FooRepo)(nil)
