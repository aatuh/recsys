package admin

import "errors"

var (
	// ErrTenantNotFound indicates the tenant does not exist.
	ErrTenantNotFound = errors.New("tenant not found")
	// ErrConfigNotFound indicates missing tenant config.
	ErrConfigNotFound = errors.New("tenant config not found")
	// ErrRulesNotFound indicates missing tenant rules.
	ErrRulesNotFound = errors.New("tenant rules not found")
	// ErrVersionMismatch indicates optimistic concurrency mismatch.
	ErrVersionMismatch = errors.New("version mismatch")
)
