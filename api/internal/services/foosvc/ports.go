package foosvc

import "context"

// Repository is the domain repository contract for Foo.
// Kept in the domain package so business logic is fully shareable.
type Repository interface {
	Create(ctx context.Context, f *Foo) error
	GetByID(ctx context.Context, id string) (*Foo, error)
	Update(ctx context.Context, f *Foo) error
	Delete(ctx context.Context, id string) error
	// Return items and total for pagination metadata.
	List(ctx context.Context, orgID, ns string,
		limit, offset int, search string) ([]Foo, int, error)
}

// Repo is kept as a type alias for compatibility across your codebase.
type Repo = Repository
