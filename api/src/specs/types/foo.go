package types

import "time"

// CreateFooDTO is the HTTP request model for Foo creation.
type CreateFooDTO struct {
	OrgID     string `json:"org_id" validate:"required"`
	Namespace string `json:"namespace" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

// UpdateFooDTO is the HTTP request model for Foo update.
type UpdateFooDTO struct {
	Name *string `json:"name" validate:"required"`
}

// FooDTO is the HTTP response model for Foo.
type FooDTO struct {
	ID        string    `json:"id" example:"01HZJ8K9M2N3P4Q5R6S7T8U9V"`
	OrgID     string    `json:"org_id" example:"org-123"`
	Namespace string    `json:"namespace" example:"default"`
	Name      string    `json:"name" example:"my-foo"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// ListMeta describes pagination metadata for list responses.
type ListMeta struct {
	Total   int                 `json:"total"`
	Count   int                 `json:"count"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	Filters map[string][]string `json:"filters,omitempty"`
	Search  string              `json:"search,omitempty"`
}

// FooListResponse is the paginated response contract for foos.
type FooListResponse struct {
	Data []FooDTO `json:"data"`
	Meta ListMeta `json:"meta"`
}
