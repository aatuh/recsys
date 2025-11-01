package explain

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Provider defines a pluggable LLM client constructor.
type Provider interface {
	Key() string
	Build(opts ProviderOptions) (LLMClient, error)
}

// ProviderOptions supply runtime configuration to provider implementations.
type ProviderOptions struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     *zap.Logger
}

// ProviderRegistry stores known explain providers.
type ProviderRegistry struct {
	providers map[string]Provider
}

// NewProviderRegistry constructs a registry with optional providers.
func NewProviderRegistry(providers ...Provider) *ProviderRegistry {
	reg := &ProviderRegistry{providers: make(map[string]Provider, len(providers))}
	for _, p := range providers {
		reg.Register(p)
	}
	return reg
}

// Register inserts or replaces a provider in the registry.
func (r *ProviderRegistry) Register(p Provider) {
	if r.providers == nil {
		r.providers = make(map[string]Provider)
	}
	r.providers[strings.ToLower(p.Key())] = p
}

// Build returns the client for the named provider or a NullClient when name is blank.
func (r *ProviderRegistry) Build(name string, opts ProviderOptions) (LLMClient, error) {
	if strings.TrimSpace(name) == "" {
		return NullClient{}, nil
	}
	if r.providers == nil {
		return nil, fmt.Errorf("no providers registered")
	}
	provider, ok := r.providers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown explain provider %q", name)
	}
	return provider.Build(opts)
}

// DefaultProviderRegistry returns the default registry with built-in providers.
func DefaultProviderRegistry() *ProviderRegistry {
	return NewProviderRegistry(openAIProvider{})
}

type openAIProvider struct{}

func (openAIProvider) Key() string { return "openai" }

func (openAIProvider) Build(opts ProviderOptions) (LLMClient, error) {
	opts.APIKey = strings.TrimSpace(opts.APIKey)
	if opts.APIKey == "" {
		return nil, fmt.Errorf("openai api key missing")
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &OpenAIClient{
		HTTP:    httpClient,
		APIKey:  opts.APIKey,
		BaseURL: strings.TrimSpace(opts.BaseURL),
		Logger:  opts.Logger,
	}, nil
}
