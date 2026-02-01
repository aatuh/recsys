package endpoints

// Admin endpoints (v1).
const (
	AdminBase             = RecsysBase + "/admin"
	AdminTenantBase       = AdminBase + "/tenants/{tenant_id}"
	AdminTenantConfig     = AdminTenantBase + "/config"
	AdminTenantRules      = AdminTenantBase + "/rules"
	AdminTenantInvalidate = AdminTenantBase + "/cache/invalidate"
	AdminTenantAudit      = AdminTenantBase + "/audit"
)
