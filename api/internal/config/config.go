package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/api-toolkit-contrib/config"
	jwtint "github.com/aatuh/api-toolkit-contrib/integrations/auth/jwt"
	"github.com/aatuh/api-toolkit-contrib/middleware/auth/devheaders"
	"github.com/aatuh/api-toolkit-contrib/telemetry"
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
	Performance PerformanceConfig
	Telemetry   TelemetryConfig
	Exposure    ExposureConfig
	Experiment  ExperimentConfig
	Explain     ExplainConfig
	Artifacts   ArtifactConfig
	Algo        AlgoConfig
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
	APIKeys         APIKeyConfig
}

// APIKeyConfig controls API key authentication.
type APIKeyConfig struct {
	Enabled    bool
	Header     string
	HashSecret string
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

// ExposureConfig configures exposure logging.
type ExposureConfig struct {
	Enabled       bool
	Path          string
	Format        string
	Fsync         bool
	RetentionDays int
	HashSalt      string
}

// ExperimentConfig configures experiment assignment.
type ExperimentConfig struct {
	Enabled         bool
	DefaultVariants []string
	Salt            string
}

// ExplainConfig controls explain/trace safeguards.
type ExplainConfig struct {
	MaxItems     int
	RequireAdmin bool
}

// ArtifactConfig controls artifact/manifest reading.
type ArtifactConfig struct {
	Enabled          bool
	ManifestTemplate string
	ManifestTTL      time.Duration
	ArtifactTTL      time.Duration
	MaxBytes         int
	S3               ArtifactS3Config
}

// ArtifactS3Config configures S3-compatible access for artifacts.
type ArtifactS3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}

// AlgoConfig configures recsys-algo defaults.
type AlgoConfig struct {
	Version                    string
	DefaultNamespace           string
	Mode                       string
	PluginEnabled              bool
	PluginPath                 string
	BlendAlpha                 float64
	BlendBeta                  float64
	BlendGamma                 float64
	ProfileBoost               float64
	ProfileWindowDays          float64
	ProfileTopNTags            int
	ProfileMinEventsForBoost   int
	ProfileColdStartMultiplier float64
	ProfileStarterBlendWeight  float64
	MMRLambda                  float64
	BrandCap                   int
	CategoryCap                int
	HalfLifeDays               float64
	CoVisWindowDays            int
	PurchasedWindowDays        int
	RuleExcludeEvents          bool
	ExcludeEventTypes          []int16
	BrandTagPrefixes           []string
	CategoryTagPrefixes        []string
	RulesEnabled               bool
	RulesRefreshInterval       time.Duration
	RulesMaxPins               int
	PopularityFanout           int
	MaxK                       int
	MaxFanout                  int
	MaxExcludeIDs              int
	MaxAnchorsInjected         int
	SessionLookbackEvents      int
	SessionLookaheadMinutes    float64
}

// PerformanceConfig groups performance-related toggles.
type PerformanceConfig struct {
	Backpressure BackpressureConfig
	Cache        CacheConfig
	PprofEnabled bool
}

// TelemetryConfig bundles observability toggles.
type TelemetryConfig struct {
	Tracing telemetry.TraceConfig
}

// BackpressureConfig controls bounded queue behavior.
type BackpressureConfig struct {
	MaxInFlight int
	MaxQueue    int
	WaitTimeout time.Duration
	RetryAfter  time.Duration
}

// CacheConfig configures TTL caches.
type CacheConfig struct {
	ConfigTTL time.Duration
	RulesTTL  time.Duration
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
		APIKeys: APIKeyConfig{
			Enabled:    loader.Bool("API_KEY_ENABLED", false),
			Header:     loader.String("API_KEY_HEADER", "X-API-Key"),
			HashSecret: loader.String("API_KEY_HASH_SECRET", ""),
		},
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
	perfCfg := PerformanceConfig{
		Backpressure: BackpressureConfig{
			MaxInFlight: loader.Int("RECSYS_BACKPRESSURE_MAX_INFLIGHT", 0),
			MaxQueue:    loader.Int("RECSYS_BACKPRESSURE_MAX_QUEUE", 0),
			WaitTimeout: loader.Duration("RECSYS_BACKPRESSURE_WAIT_TIMEOUT", 200*time.Millisecond),
			RetryAfter:  loader.Duration("RECSYS_BACKPRESSURE_RETRY_AFTER", time.Second),
		},
		Cache: CacheConfig{
			ConfigTTL: loader.Duration("RECSYS_CONFIG_CACHE_TTL", 5*time.Minute),
			RulesTTL:  loader.Duration("RECSYS_RULES_CACHE_TTL", 5*time.Minute),
		},
		PprofEnabled: loader.Bool("PPROF_ENABLED", false),
	}
	traceCfg := telemetry.LoadTraceConfig(loader)
	if strings.TrimSpace(traceCfg.ServiceName) == "" || traceCfg.ServiceName == "api" {
		traceCfg.ServiceName = "recsys-service"
	}
	exposurePath := loader.String("EXPOSURE_LOG_PATH", "")
	exposureEnabled := loader.Bool("EXPOSURE_LOG_ENABLED", exposurePath != "")
	exposureCfg := ExposureConfig{
		Enabled:       exposureEnabled,
		Path:          exposurePath,
		Format:        loader.String("EXPOSURE_LOG_FORMAT", "service_v1"),
		Fsync:         loader.Bool("EXPOSURE_LOG_FSYNC", false),
		RetentionDays: loader.Int("EXPOSURE_LOG_RETENTION_DAYS", 30),
		HashSalt:      loader.String("EXPOSURE_HASH_SALT", ""),
	}
	artifactCfg := ArtifactConfig{
		Enabled:          loader.Bool("RECSYS_ARTIFACT_MODE_ENABLED", false),
		ManifestTemplate: loader.String("RECSYS_ARTIFACT_MANIFEST_TEMPLATE", ""),
		ManifestTTL:      loader.Duration("RECSYS_ARTIFACT_MANIFEST_TTL", time.Minute),
		ArtifactTTL:      loader.Duration("RECSYS_ARTIFACT_CACHE_TTL", time.Minute),
		MaxBytes:         loader.Int("RECSYS_ARTIFACT_MAX_BYTES", 10_000_000),
		S3: ArtifactS3Config{
			Endpoint:  loader.String("RECSYS_ARTIFACT_S3_ENDPOINT", ""),
			AccessKey: loader.String("RECSYS_ARTIFACT_S3_ACCESS_KEY", ""),
			SecretKey: loader.String("RECSYS_ARTIFACT_S3_SECRET_KEY", ""),
			Region:    loader.String("RECSYS_ARTIFACT_S3_REGION", ""),
			UseSSL:    loader.Bool("RECSYS_ARTIFACT_S3_USE_SSL", false),
		},
	}
	expVariants := loader.CSV("EXPERIMENT_DEFAULT_VARIANTS")
	if len(expVariants) == 0 {
		expVariants = []string{"A", "B"}
	}
	expCfg := ExperimentConfig{
		Enabled:         loader.Bool("EXPERIMENT_ASSIGNMENT_ENABLED", true),
		DefaultVariants: expVariants,
		Salt:            loader.String("EXPERIMENT_ASSIGNMENT_SALT", exposureCfg.HashSalt),
	}
	explainCfg := ExplainConfig{
		MaxItems:     loader.Int("RECSYS_EXPLAIN_MAX_ITEMS", 50),
		RequireAdmin: loader.Bool("RECSYS_EXPLAIN_REQUIRE_ADMIN", true),
	}
	algoCfg := AlgoConfig{
		Version:                    loader.String("RECSYS_ALGO_VERSION", "recsys-algo@local"),
		DefaultNamespace:           loader.String("RECSYS_ALGO_DEFAULT_NAMESPACE", "default"),
		Mode:                       loader.String("RECSYS_ALGO_MODE", "blend"),
		PluginEnabled:              loader.Bool("RECSYS_ALGO_PLUGIN_ENABLED", false),
		PluginPath:                 loader.String("RECSYS_ALGO_PLUGIN_PATH", ""),
		BlendAlpha:                 floatEnv(loader, "RECSYS_ALGO_BLEND_ALPHA", 1),
		BlendBeta:                  floatEnv(loader, "RECSYS_ALGO_BLEND_BETA", 0),
		BlendGamma:                 floatEnv(loader, "RECSYS_ALGO_BLEND_GAMMA", 0),
		ProfileBoost:               floatEnv(loader, "RECSYS_ALGO_PROFILE_BOOST", 0.6),
		ProfileWindowDays:          floatEnv(loader, "RECSYS_ALGO_PROFILE_WINDOW_DAYS", 30),
		ProfileTopNTags:            loader.Int("RECSYS_ALGO_PROFILE_TOP_N", 5),
		ProfileMinEventsForBoost:   loader.Int("RECSYS_ALGO_PROFILE_MIN_EVENTS", 10),
		ProfileColdStartMultiplier: floatEnv(loader, "RECSYS_ALGO_PROFILE_COLD_START_MULT", 0.5),
		ProfileStarterBlendWeight:  floatEnv(loader, "RECSYS_ALGO_PROFILE_STARTER_BLEND_WEIGHT", 0.5),
		MMRLambda:                  floatEnv(loader, "RECSYS_ALGO_MMR_LAMBDA", 0),
		BrandCap:                   loader.Int("RECSYS_ALGO_BRAND_CAP", 0),
		CategoryCap:                loader.Int("RECSYS_ALGO_CATEGORY_CAP", 0),
		HalfLifeDays:               floatEnv(loader, "RECSYS_ALGO_HALF_LIFE_DAYS", 30),
		CoVisWindowDays:            loader.Int("RECSYS_ALGO_COVIS_WINDOW_DAYS", 30),
		PurchasedWindowDays:        loader.Int("RECSYS_ALGO_PURCHASED_WINDOW_DAYS", 30),
		RuleExcludeEvents:          loader.Bool("RECSYS_ALGO_RULE_EXCLUDE_EVENTS", false),
		ExcludeEventTypes:          int16CSV(loader, "RECSYS_ALGO_EXCLUDE_EVENT_TYPES"),
		BrandTagPrefixes:           loader.CSV("RECSYS_ALGO_BRAND_TAG_PREFIXES"),
		CategoryTagPrefixes:        loader.CSV("RECSYS_ALGO_CATEGORY_TAG_PREFIXES"),
		RulesEnabled:               loader.Bool("RECSYS_ALGO_RULES_ENABLED", false),
		RulesRefreshInterval:       loader.Duration("RECSYS_ALGO_RULES_REFRESH_INTERVAL", 2*time.Second),
		RulesMaxPins:               loader.Int("RECSYS_ALGO_RULES_MAX_PINS", 3),
		PopularityFanout:           loader.Int("RECSYS_ALGO_POPULARITY_FANOUT", 0),
		MaxK:                       loader.Int("RECSYS_ALGO_MAX_K", 200),
		MaxFanout:                  loader.Int("RECSYS_ALGO_MAX_FANOUT", 0),
		MaxExcludeIDs:              loader.Int("RECSYS_ALGO_MAX_EXCLUDE_IDS", 200),
		MaxAnchorsInjected:         loader.Int("RECSYS_ALGO_MAX_ANCHORS_INJECTED", 50),
		SessionLookbackEvents:      loader.Int("RECSYS_ALGO_SESSION_LOOKBACK_EVENTS", 50),
		SessionLookaheadMinutes:    floatEnv(loader, "RECSYS_ALGO_SESSION_LOOKAHEAD_MINUTES", 120),
	}
	cfg := Config{
		Config:      base,
		DocsEnabled: loader.Bool("DOCS_ENABLED", !config.IsProduction(base.Env)),
		Health:      health.LoadConfig(loader),
		App: AppConfig{
			BaseURL:    loader.String("FRONTEND_BASE_URL", "http://localhost:3000"),
			ProjectTag: loader.String("PROJECT_TAG", ""),
		},
		Auth:        authCfg,
		RateLimit:   rateCfg,
		Audit:       auditCfg,
		Performance: perfCfg,
		Telemetry: TelemetryConfig{
			Tracing: traceCfg,
		},
		Exposure:   exposureCfg,
		Experiment: expCfg,
		Explain:    explainCfg,
		Artifacts:  artifactCfg,
		Algo:       algoCfg,
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

func floatEnv(loader *config.Loader, key string, def float64) float64 {
	if loader == nil {
		return def
	}
	raw := strings.TrimSpace(loader.String(key, ""))
	if raw == "" {
		return def
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return val
}

func int16CSV(loader *config.Loader, key string) []int16 {
	if loader == nil {
		return nil
	}
	values := loader.CSV(key)
	if len(values) == 0 {
		return nil
	}
	out := make([]int16, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		val, err := strconv.ParseInt(raw, 10, 16)
		if err != nil {
			continue
		}
		out = append(out, int16(val))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
