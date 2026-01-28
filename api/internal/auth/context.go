package auth

import "context"

type ctxKey struct{}

// TenantSource describes how a tenant identifier was resolved.
type TenantSource string

const (
	TenantSourceClaim TenantSource = "claim"
	TenantSourceDev   TenantSource = "dev"
)

// Info captures authentication context derived from the request.
type Info struct {
	UserID       string
	TenantID     string
	TenantSource TenantSource
	Roles        []string
}

// WithInfo stores auth info in context.
func WithInfo(ctx context.Context, info Info) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey{}, info)
}

// FromContext retrieves auth info from context.
func FromContext(ctx context.Context) (Info, bool) {
	if ctx == nil {
		return Info{}, false
	}
	info, ok := ctx.Value(ctxKey{}).(Info)
	return info, ok
}

// TenantIDFromContext returns the tenant id from auth info.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	info, ok := FromContext(ctx)
	if !ok || info.TenantID == "" {
		return "", false
	}
	return info.TenantID, true
}

// UserIDFromContext returns the user id from auth info.
func UserIDFromContext(ctx context.Context) (string, bool) {
	info, ok := FromContext(ctx)
	if !ok || info.UserID == "" {
		return "", false
	}
	return info.UserID, true
}

// RolesFromContext returns the role list from auth info.
func RolesFromContext(ctx context.Context) []string {
	info, ok := FromContext(ctx)
	if !ok || len(info.Roles) == 0 {
		return nil
	}
	out := make([]string, len(info.Roles))
	copy(out, info.Roles)
	return out
}
