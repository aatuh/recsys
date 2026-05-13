package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/aatuh/recsys-suite/api/internal/auth"
)

func TestAdminHandlerAuditDetailPolicy(t *testing.T) {
	t.Parallel()

	handler := NewAdminHandler(nil, nil, nil, WithAuditDetailRoles("operator", "admin"))
	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{name: "viewer", roles: []string{"viewer"}, want: false},
		{name: "operator", roles: []string{"operator"}, want: true},
		{name: "admin", roles: []string{"admin"}, want: true},
		{name: "case insensitive", roles: []string{"Admin"}, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest("GET", "/v1/admin/tenants/demo/audit", nil)
			req = req.WithContext(auth.WithInfo(req.Context(), auth.Info{Roles: tc.roles}))
			if got := handler.canReadAuditDetails(req); got != tc.want {
				t.Fatalf("canReadAuditDetails() = %v, want %v", got, tc.want)
			}
		})
	}
}
