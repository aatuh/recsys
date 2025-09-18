package explain

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubCollector struct {
	facts    FactsPack
	evidence []Evidence
	err      error
}

func (s *stubCollector) Collect(ctx context.Context, orgID uuid.UUID, req Request) (FactsPack, []Evidence, error) {
	return s.facts, s.evidence, s.err
}

type stubClient struct {
	resp string
	err  error
}

func (s stubClient) Generate(ctx context.Context, model string, systemPrompt string, userPrompt string, maxTokens int) (string, error) {
	return s.resp, s.err
}

func TestServiceFallbackCreatesStructuredMarkdown(t *testing.T) {
	facts := FactsPack{}
	facts.Target.Type = "item"
	facts.Target.ID = "slot-1"
	facts.Window.From = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	facts.Window.To = time.Now().UTC().Format(time.RFC3339)
	facts.Context.Namespace = "default"
	facts.Context.Surface = "home"
	facts.Metrics.Impressions = 0
	facts.Metrics.Clicks = 0

	collector := &stubCollector{facts: facts}
	svc := &Service{
		Collector: collector,
		Client:    NullClient{},
		Config: Config{
			Enabled:   false,
			CacheTTL:  time.Minute,
			Timeout:   time.Second,
			MaxTokens: 100,
		},
	}

	orgID := uuid.New()
	req := Request{
		TargetType: "ITEM",
		TargetID:   "slot-1",
		Namespace:  "default",
		Surface:    "home",
		From:       time.Now().Add(-time.Hour),
		To:         time.Now(),
	}

	result, err := svc.Explain(context.Background(), orgID, req)
	require.NoError(t, err)
	require.Contains(t, result.Markdown, "## Summary")
	require.Contains(t, result.Markdown, "## Status")
	require.Contains(t, result.Markdown, "## Key findings")
	require.Contains(t, result.Markdown, "## Likely causes")
	require.Contains(t, result.Markdown, "## Suggested fix")
	require.Equal(t, "fallback", result.Model)
}

func TestCacheHit(t *testing.T) {
	facts := FactsPack{}
	facts.Target.Type = "item"
	facts.Target.ID = "slot-2"
	facts.Window.From = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	facts.Window.To = time.Now().UTC().Format(time.RFC3339)
	facts.Context.Namespace = "default"
	facts.Context.Surface = "home"

	collector := &stubCollector{facts: facts}
	svc := &Service{
		Collector: collector,
		Client:    stubClient{resp: "analysis"},
		Config: Config{
			Enabled:      true,
			CacheTTL:     time.Minute,
			Timeout:      time.Second,
			MaxTokens:    100,
			ModelPrimary: "o4-mini",
		},
	}

	orgID := uuid.New()
	req := Request{
		TargetType: "ITEM",
		TargetID:   "slot-2",
		Namespace:  "default",
		Surface:    "home",
		From:       time.Now().Add(-time.Hour),
		To:         time.Now(),
	}

	_, err := svc.Explain(context.Background(), orgID, req)
	require.NoError(t, err)
	result, err := svc.Explain(context.Background(), orgID, req)
	require.NoError(t, err)
	require.True(t, result.CacheHit)
}
