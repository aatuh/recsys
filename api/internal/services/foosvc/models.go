package foosvc

import "time"

// Foo is a minimal example entity.
type Foo struct {
	ID        string
	OrgID     string
	Namespace string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateInput is validated input for creating Foo.
type CreateInput struct {
	OrgID     string
	Namespace string
	Name      string
}

// UpdateInput is validated input for updating Foo.
type UpdateInput struct {
	ID   string
	Name *string
}
