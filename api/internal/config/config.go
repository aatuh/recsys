package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aatuh/api-toolkit-contrib/config"
	jwtint "github.com/aatuh/api-toolkit-contrib/integrations/auth/jwt"
	"github.com/aatuh/api-toolkit-contrib/middleware/auth/devheaders"
	"github.com/aatuh/api-toolkit/endpoints/health"
)

// Config extends the base toolkit config with app-specific settings.
type Config struct {
	config.Config
	DocsEnabled bool
	Health      health.Config
	App         AppConfig
	Auth        AuthConfig
	RateLimit   RateLimitConfig
	Audit       AuditConfig
}

// AppConfig holds app-wide public-facing configuration.
type AppConfig struct {
	BaseURL    string
	ProjectTag string
}

// AuthConfig holds authentication and tenancy settings.
type AuthConfig struct {
	Required           bool
	RequireTenantClaim bool
	TenantHeader       string
	TenantClaimKeys    []string
	RoleClaimKeys      []string
	AdminRole          string
	JWKSAllowedHosts   []string
	AllowInsecureJWKS  bool

	JWT             jwtint.Config
	DevHeaders      devheaders.Config
	DevTenantHeader string
}

// RateLimitConfig configures per-tenant rate limiting.
type RateLimitConfig struct {
	TenantEnabled    bool
	TenantCapacity   float64
	TenantRefillRate float64
	TenantRetryAfter time.Duration
	TenantFailOpen   bool
}

// AuditConfig configures audit logging.
type AuditConfig struct {
	Enabled bool
	Path    string
	Fsync   bool
}

// Load loads env vars into Config structure.
func Load() Config {
	if err := applySecretFile("DATABASE_URL"); err != nil {
		panic(err)
	}
	loader := config.NewLoader()
	base, err := config.LoadFromEnv(loader)
	if err != nil {
		panic(err)
	}
	jwtCfg := jwtint.LoadConfig(loader)
	devCfg := devheaders.Config{
		Enabled:         loader.Bool("DEV_AUTH_ENABLED", false),
		UserIDHeader:    loader.String("DEV_AUTH_USER_ID_HEADER", "X-Dev-User-Id"),
		EmailHeader:     loader.String("DEV_AUTH_EMAIL_HEADER", "X-Dev-User-Email"),
		FirstNameHeader: loader.String("DEV_AUTH_FIRST_NAME_HEADER", "X-Dev-User-First"),
		LastNameHeader:  loader.String("DEV_AUTH_LAST_NAME_HEADER", "X-Dev-User-Last"),
		DefaultLanguage: loader.String("DEV_AUTH_DEFAULT_LANGUAGE", ""),
	}
	tenantClaims := loader.CSV("AUTH_TENANT_CLAIMS")
	if len(tenantClaims) == 0 {
		tenantClaims = []string{"tenant_id", "org_id", "orgId", "organization_id"}
	}
	roleClaims := loader.CSV("AUTH_ROLE_CLAIMS")
	if len(roleClaims) == 0 {
		roleClaims = []string{"roles", "role", "scope", "scopes"}
	}
	requireTenantClaimDefault := config.IsProduction(base.Env)
	authCfg := AuthConfig{
		Required:           loader.Bool("AUTH_REQUIRED", true),
		RequireTenantClaim: loader.Bool("AUTH_REQUIRE_TENANT_CLAIM", requireTenantClaimDefault),
		TenantHeader:       loader.String("TENANT_HEADER_NAME", "X-Org-Id"),
		TenantClaimKeys:    tenantClaims,
		RoleClaimKeys:      roleClaims,
		AdminRole:          loader.String("AUTH_ADMIN_ROLE", "admin"),
		JWKSAllowedHosts:   loader.CSV("AUTH_JWKS_ALLOWED_HOSTS"),
		AllowInsecureJWKS:  loader.Bool("AUTH_ALLOW_INSECURE_JWKS", false),
		JWT:                jwtCfg,
		DevHeaders:         devCfg,
		DevTenantHeader:    loader.String("DEV_AUTH_TENANT_HEADER", "X-Dev-Org-Id"),
	}
	rateCfg := RateLimitConfig{
		TenantEnabled:    loader.Bool("TENANT_RATE_LIMIT_ENABLED", true),
		TenantCapacity:   float64(loader.Int("TENANT_RATE_LIMIT_CAPACITY", 60)),
		TenantRefillRate: float64(loader.Int("TENANT_RATE_LIMIT_REFILL_RATE", 30)),
		TenantRetryAfter: loader.Duration("TENANT_RATE_LIMIT_RETRY_AFTER", time.Second),
		TenantFailOpen:   loader.Bool("TENANT_RATE_LIMIT_FAIL_OPEN", false),
	}
	auditPath := loader.String("AUDIT_LOG_PATH", "")
	auditCfg := AuditConfig{
		Enabled: loader.Bool("AUDIT_LOG_ENABLED", auditPath != ""),
		Path:    auditPath,
		Fsync:   loader.Bool("AUDIT_LOG_FSYNC", false),
	}
	cfg := Config{
		Config:      base,
		DocsEnabled: loader.Bool("DOCS_ENABLED", !config.IsProduction(base.Env)),
		Health:      health.LoadConfig(loader),
		App: AppConfig{
			BaseURL:    loader.String("FRONTEND_BASE_URL", "http://localhost:3000"),
			ProjectTag: loader.String("PROJECT_TAG", ""),
		},
		Auth:      authCfg,
		RateLimit: rateCfg,
		Audit:     auditCfg,
	}
	if err := loader.Err(); err != nil {
		panic(err)
	}
	return cfg
}

func applySecretFile(envKey string) error {
	path := strings.TrimSpace(os.Getenv(envKey + "_FILE"))
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s_FILE: %w", envKey, err)
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return fmt.Errorf("%s_FILE is empty", envKey)
	}
	return os.Setenv(envKey, value)
}
