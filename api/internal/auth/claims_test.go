package auth

import "testing"

func TestExtractTenantNestedClaim(t *testing.T) {
	claims := map[string]any{
		"app_metadata": map[string]any{
			"tenant_id": "demo",
		},
	}
	tenant := ExtractTenant(claims, []string{"app_metadata.tenant_id"})
	if tenant != "demo" {
		t.Fatalf("expected tenant demo, got %q", tenant)
	}
}

func TestExtractRolesNestedClaim(t *testing.T) {
	claims := map[string]any{
		"realm_access": map[string]any{
			"roles": []any{"viewer", "operator"},
		},
	}
	roles := ExtractRoles(claims, []string{"realm_access.roles"})
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %v", roles)
	}
}
