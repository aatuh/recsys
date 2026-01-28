package endpoints

// Keep all endpoint constants under specs/endpoints.
const (
	FooBase   = "/api/v1/foo"
	FooByID   = FooBase + "/{id}"
	FooList   = FooBase
	FooCreate = FooBase
	FooUpdate = FooByID
	FooDelete = FooByID
)
