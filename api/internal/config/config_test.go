package config

import (
	"strings"
	"testing"

	toolkitconfig "github.com/aatuh/api-toolkit/contrib/v2/config"
)

func TestValidateAllowsProductionWhenRequiredSecretsConfigured(t *testing.T) {
	cfg := Config{
		Config: configBase("production"),
		Auth: AuthConfig{
			APIKeys: APIKeyConfig{
				Enabled:    true,
				HashSecret: "api-key-pepper",
			},
		},
		Exposure: ExposureConfig{
			Enabled:  true,
			HashSalt: "exposure-salt",
		},
		Experiment: ExperimentConfig{
			Enabled: true,
			Salt:    "assignment-salt",
		},
		Performance: PerformanceConfig{
			PprofEnabled: true,
		},
	}
	cfg.Addr = "127.0.0.1:8000"

	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsProductionSecretsWhenFeatureEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "exposure logging without salt",
			cfg: Config{
				Config:   configBase("production"),
				Exposure: ExposureConfig{Enabled: true},
			},
			want: "EXPOSURE_HASH_SALT",
		},
		{
			name: "experiment assignment without salt",
			cfg: Config{
				Config:     configBase("production"),
				Experiment: ExperimentConfig{Enabled: true},
			},
			want: "EXPERIMENT_ASSIGNMENT_SALT",
		},
		{
			name: "api key auth without hash secret",
			cfg: Config{
				Config: configBase("production"),
				Auth: AuthConfig{
					APIKeys: APIKeyConfig{Enabled: true},
				},
			},
			want: "API_KEY_HASH_SECRET",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.cfg)
			if err == nil {
				t.Fatal("Validate() error = nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate() error = %q, want it to mention %q", err.Error(), tc.want)
			}
			if strings.Contains(err.Error(), "api-key-pepper") ||
				strings.Contains(err.Error(), "exposure-salt") ||
				strings.Contains(err.Error(), "assignment-salt") {
				t.Fatalf("Validate() leaked a secret value: %q", err.Error())
			}
		})
	}
}

func TestValidateRejectsProductionS3WithoutTLS(t *testing.T) {
	cfg := Config{
		Config:    configBase("production"),
		Artifacts: ArtifactConfig{Enabled: true, S3: ArtifactS3Config{Endpoint: "s3.example.com", UseSSL: false}},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() error = nil")
	}
	if !strings.Contains(err.Error(), "RECSYS_ARTIFACT_S3_USE_SSL") {
		t.Fatalf("Validate() error = %q, want S3 TLS guidance", err.Error())
	}
}

func TestValidateAllowsLocalS3WithoutTLS(t *testing.T) {
	cfg := Config{
		Config:    configBase("development"),
		Artifacts: ArtifactConfig{Enabled: true, S3: ArtifactS3Config{Endpoint: "minio:9000", UseSSL: false}},
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateAllowsMissingSecretsOutsideProduction(t *testing.T) {
	cfg := Config{
		Config:     configBase("development"),
		Auth:       AuthConfig{APIKeys: APIKeyConfig{Enabled: true}},
		Exposure:   ExposureConfig{Enabled: true},
		Experiment: ExperimentConfig{Enabled: true},
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidExperimentDefinitions(t *testing.T) {
	traffic := 150.0
	cfg := Config{
		Config: configBase("development"),
		Experiment: ExperimentConfig{
			Definitions: []ExperimentDefinition{
				{ID: "exp-1", Enabled: true, TrafficPercent: &traffic, Variants: []string{"A", "B"}},
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() error = nil")
	}
	if !strings.Contains(err.Error(), "traffic_percent") {
		t.Fatalf("Validate() error = %q, want traffic_percent", err.Error())
	}
}

func TestParseExperimentDefinitionsAppliesValidation(t *testing.T) {
	defs, err := parseExperimentDefinitions(`[{
		"id": "exp-home",
		"enabled": true,
		"surface": "home",
		"traffic_percent": 25,
		"variants": ["A", "B"],
		"starts_at": "2026-01-01T00:00:00Z",
		"ends_at": "2026-02-01T00:00:00Z"
	}]`)
	if err != nil {
		t.Fatalf("parseExperimentDefinitions() error = %v", err)
	}
	if len(defs) != 1 || defs[0].ID != "exp-home" || defs[0].TrafficPercent == nil || *defs[0].TrafficPercent != 25 {
		t.Fatalf("definitions = %+v, want parsed experiment", defs)
	}

	_, err = parseExperimentDefinitions(`[{"id":"exp-home","variants":["A","A"]}]`)
	if err == nil {
		t.Fatal("parseExperimentDefinitions() duplicate variants error = nil")
	}
	if !strings.Contains(err.Error(), "duplicate experiment variant") {
		t.Fatalf("parseExperimentDefinitions() error = %q, want duplicate variant", err.Error())
	}
}

func TestFloatEnvPanicsOnInvalidValue(t *testing.T) {
	t.Setenv("RECSYS_TEST_FLOAT", "not-a-float")

	defer func() {
		if recover() == nil {
			t.Fatal("floatEnv() did not panic")
		}
	}()

	_ = floatEnv(toolkitconfig.NewLoader(), "RECSYS_TEST_FLOAT", 1)
}

func TestFloatEnvUsesDefaultWhenUnset(t *testing.T) {
	if got := floatEnv(toolkitconfig.NewLoader(), "RECSYS_TEST_FLOAT_UNSET", 1.5); got != 1.5 {
		t.Fatalf("floatEnv() = %v, want 1.5", got)
	}
}

func TestInt16CSVPanicsOnInvalidValue(t *testing.T) {
	t.Setenv("RECSYS_TEST_INT16S", "1,bad")

	defer func() {
		if recover() == nil {
			t.Fatal("int16CSV() did not panic")
		}
	}()

	_ = int16CSV(toolkitconfig.NewLoader(), "RECSYS_TEST_INT16S")
}

func TestInt16CSVParsesValidValues(t *testing.T) {
	t.Setenv("RECSYS_TEST_INT16S", "1, 2")

	got := int16CSV(toolkitconfig.NewLoader(), "RECSYS_TEST_INT16S")
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("int16CSV() = %#v", got)
	}
}

func TestValidateRejectsPprofOnPublicBind(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{name: "wildcard", addr: ":8000"},
		{name: "all ipv4", addr: "0.0.0.0:8000"},
		{name: "all ipv6", addr: "[::]:8000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				Config:      configBase("development"),
				Performance: PerformanceConfig{PprofEnabled: true},
			}
			cfg.Addr = tc.addr

			err := Validate(cfg)
			if err == nil {
				t.Fatal("Validate() error = nil")
			}
			if !strings.Contains(err.Error(), "PPROF_ENABLED") {
				t.Fatalf("Validate() error = %q, want pprof guidance", err.Error())
			}
		})
	}
}

func TestValidateAllowsPprofOnLoopbackBind(t *testing.T) {
	for _, addr := range []string{"localhost:8000", "127.0.0.1:8000", "[::1]:8000"} {
		t.Run(addr, func(t *testing.T) {
			cfg := Config{
				Config:      configBase("development"),
				Performance: PerformanceConfig{PprofEnabled: true},
			}
			cfg.Addr = addr

			if err := Validate(cfg); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func configBase(env string) toolkitconfig.Config {
	return toolkitconfig.Config{
		Addr:        "127.0.0.1:8000",
		DatabaseURL: "postgres://user:pass@localhost:5432/db",
		LogLevel:    "info",
		Env:         env,
	}
}
