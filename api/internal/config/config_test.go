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
