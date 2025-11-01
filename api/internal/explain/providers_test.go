package explain

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultRegistryBuildsOpenAI(t *testing.T) {
	reg := DefaultProviderRegistry()
	client, err := reg.Build("openai", ProviderOptions{
		APIKey:     "test-key",
		BaseURL:    "",
		HTTPClient: &http.Client{Timeout: time.Second},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*OpenAIClient); !ok {
		t.Fatalf("expected openai client, got %T", client)
	}
}

func TestRegistryReturnsErrorForUnknownProvider(t *testing.T) {
	reg := DefaultProviderRegistry()
	if _, err := reg.Build("bogus", ProviderOptions{}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRegistryReturnsNullClientForBlankProvider(t *testing.T) {
	reg := DefaultProviderRegistry()
	client, err := reg.Build("", ProviderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(NullClient); !ok {
		t.Fatalf("expected NullClient, got %T", client)
	}
}
