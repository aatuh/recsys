package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"recsys/internal/types"
)

// Config represents the fully parsed application configuration.
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Debug          DebugConfig
	HTTP           HTTPConfig
	Auth           AuthConfig
	Recommendation RecommendationConfig
	Rules          RulesConfig
	Audit          AuditConfig
	Explain        ExplainConfig
	Migrations     MigrationConfig
}

type ServerConfig struct {
	Port              string
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type DatabaseConfig struct {
	URL         string
	MaxConnIdle time.Duration
	MinConns    int32
	MaxConns    int32
}

type DebugConfig struct {
	Environment string
	AppDebug    bool
}

type HTTPConfig struct {
	CORS CORSConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

type AuthConfig struct {
	Enabled   bool
	APIKeys   map[string]APIKeyConfig
	RateLimit RateLimitConfig
}

type APIKeyConfig struct {
	AllowAll bool
	OrgIDs   []uuid.UUID
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
	Burst             int
}

type RecommendationConfig struct {
	DefaultOrgID        uuid.UUID
	HalfLifeDays        float64
	CoVisWindowDays     float64
	PopularityFanout    int
	MMRLambda           float64
	BrandCap            int
	CategoryCap         int
	RuleExcludeEvents   bool
	ExcludeEventTypes   []int16
	BrandTagPrefixes    []string
	CategoryTagPrefixes []string
	PurchasedWindowDays float64
	Profile             ProfileConfig
	Blend               BlendConfig
	BanditAlgo          types.Algorithm
}

type ProfileConfig struct {
	WindowDays float64
	Boost      float64
	TopNTags   int
}

type BlendConfig struct {
	Alpha float64
	Beta  float64
	Gamma float64
}

type RulesConfig struct {
	Enabled      bool
	CacheRefresh time.Duration
	MaxPinSlots  int
	AuditSample  float64
}

type AuditConfig struct {
	DecisionTrace DecisionTraceConfig
}

type DecisionTraceConfig struct {
	Enabled         bool
	QueueSize       int
	BatchSize       int
	FlushInterval   time.Duration
	SampleDefault   float64
	NamespaceSample map[string]float64
	Salt            string
}

type ExplainConfig struct {
	Enabled       bool
	Provider      string
	ModelPrimary  string
	ModelEscalate string
	Timeout       time.Duration
	MaxTokens     int
	APIKey        string
	BaseURL       string
}

type MigrationConfig struct {
	RunOnStart bool
	Dir        string
}

// Load reads configuration from the provided source. If src is nil, the
// process environment is used.
func Load(ctx context.Context, src Source) (Config, error) {
	_ = ctx // reserved for future use (e.g., secret managers)
	l := newLoader(src)

	var cfg Config

	cfg.Server = ServerConfig{
		Port:              l.requiredString("API_PORT"),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	cfg.Migrations = MigrationConfig{
		RunOnStart: l.bool("MIGRATE_ON_START", false),
		Dir:        l.optionalString("MIGRATIONS_DIR", "migrations"),
	}

	if cfg.Migrations.RunOnStart && cfg.Migrations.Dir == "" {
		l.appendErr("MIGRATIONS_DIR", fmt.Errorf("must be set when MIGRATE_ON_START=true"))
	}

	cfg.Database = DatabaseConfig{
		URL:         l.requiredString("DATABASE_URL"),
		MaxConnIdle: 90 * time.Second,
		MinConns:    0,
		MaxConns:    10,
	}

	cfg.Debug = DebugConfig{
		Environment: strings.ToLower(l.optionalString("ENV", "development")),
		AppDebug:    l.bool("APP_DEBUG", false),
	}

	cfg.HTTP = HTTPConfig{
		CORS: CORSConfig{
			AllowedOrigins:   l.stringSlice("CORS_ALLOWED_ORIGINS", ',', true),
			AllowCredentials: l.bool("CORS_ALLOW_CREDENTIALS", false),
		},
	}

	cfg.Auth = AuthConfig{
		Enabled: l.bool("API_AUTH_ENABLED", false),
		APIKeys: make(map[string]APIKeyConfig),
	}

	rawAPIKeys := l.optionalString("API_AUTH_KEYS", "")
	if rawAPIKeys != "" {
		entries := strings.Split(rawAPIKeys, ",")
		for _, entry := range entries {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			parts := strings.SplitN(entry, ":", 2)
			if len(parts) != 2 {
				l.appendErr("API_AUTH_KEYS", fmt.Errorf("invalid entry %q (expected key:org1|org2 or key:*)", entry))
				continue
			}
			key := strings.TrimSpace(parts[0])
			if key == "" {
				l.appendErr("API_AUTH_KEYS", fmt.Errorf("missing api key identifier in entry %q", entry))
				continue
			}
			if _, exists := cfg.Auth.APIKeys[key]; exists {
				l.appendErr("API_AUTH_KEYS", fmt.Errorf("duplicate api key %q", key))
				continue
			}
			scope := strings.TrimSpace(parts[1])
			if scope == "" {
				l.appendErr("API_AUTH_KEYS", fmt.Errorf("missing scope definition for key %q", key))
				continue
			}
			var keyCfg APIKeyConfig
			valid := true
			if scope == "*" {
				keyCfg.AllowAll = true
			} else {
				segments := strings.Split(scope, "|")
				seen := make(map[uuid.UUID]struct{})
				for _, seg := range segments {
					seg = strings.TrimSpace(seg)
					if seg == "" {
						continue
					}
					id, err := uuid.Parse(seg)
					if err != nil {
						l.appendErr("API_AUTH_KEYS", fmt.Errorf("invalid org id %q for key %s", seg, key))
						valid = false
						continue
					}
					if _, exists := seen[id]; exists {
						continue
					}
					seen[id] = struct{}{}
					keyCfg.OrgIDs = append(keyCfg.OrgIDs, id)
				}
				if valid && len(keyCfg.OrgIDs) == 0 {
					l.appendErr("API_AUTH_KEYS", fmt.Errorf("no valid org ids for key %q", key))
					valid = false
				}
			}
			if !valid {
				continue
			}
			cfg.Auth.APIKeys[key] = keyCfg
		}
	}

	if cfg.Auth.Enabled && len(cfg.Auth.APIKeys) == 0 {
		l.appendErr("API_AUTH_KEYS", fmt.Errorf("must define at least one api key when API_AUTH_ENABLED=true"))
	}

	cfg.Auth.RateLimit = RateLimitConfig{
		RequestsPerMinute: 600,
		Burst:             60,
	}

	if rpmStr, ok := l.lookup("API_RATE_LIMIT_RPM"); ok {
		if rpmStr == "" {
			cfg.Auth.RateLimit.RequestsPerMinute = 0
		} else {
			rpm, err := strconv.Atoi(rpmStr)
			if err != nil || rpm < 0 {
				l.appendErr("API_RATE_LIMIT_RPM", fmt.Errorf("must be an integer >= 0"))
				cfg.Auth.RateLimit.RequestsPerMinute = 0
			} else {
				cfg.Auth.RateLimit.RequestsPerMinute = rpm
			}
		}
	}

	if cfg.Auth.RateLimit.RequestsPerMinute > 0 {
		if burstStr, ok := l.lookup("API_RATE_LIMIT_BURST"); ok {
			if burstStr == "" {
				cfg.Auth.RateLimit.Burst = 0
			} else {
				burst, err := strconv.Atoi(burstStr)
				if err != nil || burst <= 0 {
					l.appendErr("API_RATE_LIMIT_BURST", fmt.Errorf("must be an integer > 0"))
					cfg.Auth.RateLimit.Burst = 0
				} else {
					cfg.Auth.RateLimit.Burst = burst
				}
			}
		}
		if cfg.Auth.RateLimit.Burst <= 0 {
			cfg.Auth.RateLimit.Burst = 60
		}
		cfg.Auth.RateLimit.Enabled = true
	} else {
		cfg.Auth.RateLimit.Burst = 0
		cfg.Auth.RateLimit.Enabled = false
	}

	cfg.Recommendation = RecommendationConfig{}

	if org := l.requiredString("ORG_ID"); org != "" {
		id, err := uuid.Parse(org)
		if err != nil {
			l.appendErr("ORG_ID", fmt.Errorf("invalid uuid: %w", err))
		} else {
			cfg.Recommendation.DefaultOrgID = id
		}
	}

	cfg.Recommendation.HalfLifeDays = l.positiveFloat("POPULARITY_HALFLIFE_DAYS")
	cfg.Recommendation.CoVisWindowDays = l.positiveFloat("COVIS_WINDOW_DAYS")
	cfg.Recommendation.PopularityFanout = l.intGreaterThan("POPULARITY_FANOUT", 0)
	cfg.Recommendation.MMRLambda = l.floatBetween("MMR_LAMBDA", 0, 1)
	cfg.Recommendation.BrandCap = l.intNonNegative("BRAND_CAP")
	cfg.Recommendation.CategoryCap = l.intNonNegative("CATEGORY_CAP")
	cfg.Recommendation.RuleExcludeEvents = l.bool("RULE_EXCLUDE_EVENTS", false)
	cfg.Recommendation.ExcludeEventTypes = l.optionalIntSlice("EXCLUDE_EVENT_TYPES")
	cfg.Recommendation.BrandTagPrefixes = l.stringSlice("BRAND_TAG_PREFIXES", ',', false)
	cfg.Recommendation.CategoryTagPrefixes = l.stringSlice("CATEGORY_TAG_PREFIXES", ',', false)
	cfg.Recommendation.PurchasedWindowDays = l.positiveFloat("PURCHASED_WINDOW_DAYS")

	cfg.Recommendation.Profile = ProfileConfig{
		WindowDays: l.positiveFloat("PROFILE_WINDOW_DAYS"),
		Boost:      l.nonNegativeFloat("PROFILE_BOOST"),
		TopNTags:   l.intGreaterThan("PROFILE_TOP_N", 0),
	}

	cfg.Recommendation.Blend = BlendConfig{
		Alpha: l.nonNegativeFloat("BLEND_ALPHA"),
		Beta:  l.nonNegativeFloat("BLEND_BETA"),
		Gamma: l.nonNegativeFloat("BLEND_GAMMA"),
	}

	if algo := l.requiredString("BANDIT_ALGO"); algo != "" {
		parsed, err := types.ParseAlgorithm(algo)
		if err != nil {
			l.appendErr("BANDIT_ALGO", err)
		} else {
			cfg.Recommendation.BanditAlgo = parsed
		}
	}

	cfg.Rules = RulesConfig{
		Enabled:      l.bool("RULES_ENABLE", false),
		CacheRefresh: l.optionalDuration("RULES_CACHE_REFRESH", 2*time.Second),
		MaxPinSlots:  l.optionalIntGreaterThan("RULES_MAX_PIN_SLOTS", 0, 3),
		AuditSample:  l.optionalPositiveFloat("RULES_AUDIT_SAMPLE", 1.0),
	}

	cfg.Audit = AuditConfig{}
	decisionTraceEnabled := l.bool("AUDIT_DECISIONS_ENABLED", false)
	if decisionTraceEnabled {
		cfg.Audit.DecisionTrace = DecisionTraceConfig{
			Enabled:         true,
			QueueSize:       l.intGreaterThan("AUDIT_DECISIONS_QUEUE", 0),
			BatchSize:       l.intGreaterThan("AUDIT_DECISIONS_BATCH", 0),
			FlushInterval:   l.requiredDuration("AUDIT_DECISIONS_FLUSH_INTERVAL"),
			SampleDefault:   l.floatBetween("AUDIT_DECISIONS_SAMPLE_DEFAULT", 0, 1),
			NamespaceSample: l.optionalStringMap("AUDIT_DECISIONS_SAMPLE_OVERRIDES"),
			Salt:            l.requiredString("AUDIT_DECISIONS_SALT"),
		}
	}

	cfg.Explain = ExplainConfig{}
	cfg.Explain.Enabled = l.bool("LLM_EXPLAIN_ENABLED", false)
	cfg.Explain.Provider = l.optionalString("LLM_PROVIDER", "")
	cfg.Explain.ModelPrimary = l.optionalString("LLM_MODEL_PRIMARY", "")
	cfg.Explain.ModelEscalate = l.optionalString("LLM_MODEL_ESCALATE", "")
	cfg.Explain.APIKey = l.optionalString("LLM_API_KEY", "")
	cfg.Explain.BaseURL = l.optionalString("LLM_BASE_URL", "")
	cfg.Explain.Timeout = l.optionalDuration("LLM_TIMEOUT", 6*time.Second)
	cfg.Explain.MaxTokens = l.optionalIntGreaterThan("LLM_MAX_TOKENS", 0, 1200)

	if err := l.err(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
