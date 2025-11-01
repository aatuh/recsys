package explain

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service orchestrates facts collection, prompting, LLM calls, and caching.
type FactsCollector interface {
	Collect(ctx context.Context, orgID uuid.UUID, req Request) (FactsPack, []Evidence, error)
}

type Service struct {
	Collector FactsCollector
	Client    LLMClient
	Config    Config
	Logger    *zap.Logger

	cacheMu sync.RWMutex
	cache   map[string]cacheEntry
}

type cacheEntry struct {
	result  Result
	expires time.Time
}

// LLMClient abstracts the model invocation.
type LLMClient interface {
	Generate(ctx context.Context, model string, systemPrompt string, userPrompt string, maxTokens int) (string, error)
}

// Explain produces the markdown explanation and facts.
func (s *Service) Explain(ctx context.Context, orgID uuid.UUID, req Request) (Result, error) {
	if s.cache == nil {
		s.cache = make(map[string]cacheEntry)
	}

	facts, evidence, err := s.Collector.Collect(ctx, orgID, req)
	if err != nil {
		return Result{}, err
	}

	factsJSON, err := json.Marshal(facts)
	if err != nil {
		return Result{}, err
	}

	key := cacheKey(req, factsJSON)
	if hit, ok := s.lookupCache(key); ok {
		hit.CacheHit = true
		hit.Facts = facts
		return hit, nil
	}

	model := s.selectModel(len(factsJSON))
	systemPrompt := defaultSystemPrompt
	userPrompt := renderUserPrompt(req, string(factsJSON))

	var markdown string
	var warnings []string

	if !s.Config.Enabled || s.Client == nil {
		markdown = fallbackMarkdown(facts, evidence)
		warnings = append(warnings, "llm_disabled")
	} else {
		ctx, cancel := context.WithTimeout(ctx, s.Config.Timeout)
		defer cancel()
		markdown, err = s.Client.Generate(ctx, model, systemPrompt, userPrompt, s.Config.MaxTokens)
		if err != nil {
			markdown = fallbackMarkdown(facts, evidence)
			if errors.Is(err, ErrCircuitOpen) {
				warnings = append(warnings, "llm_circuit_open")
			} else {
				warnings = append(warnings, fmt.Sprintf("llm_error:%v", err))
			}
		}
	}

	result := Result{
		Markdown: markdown,
		Facts:    facts,
		Model:    model,
		CacheHit: false,
		Warnings: warnings,
	}

	s.storeCache(key, result)
	return result, nil
}

func (s *Service) lookupCache(key string) (Result, bool) {
	if s.Config.CacheTTL <= 0 {
		return Result{}, false
	}
	s.cacheMu.RLock()
	entry, ok := s.cache[key]
	s.cacheMu.RUnlock()
	if !ok || time.Now().After(entry.expires) {
		if ok {
			s.cacheMu.Lock()
			delete(s.cache, key)
			s.cacheMu.Unlock()
		}
		return Result{}, false
	}
	return entry.result, true
}

func (s *Service) storeCache(key string, result Result) {
	if s.Config.CacheTTL <= 0 {
		return
	}
	s.cacheMu.Lock()
	s.cache[key] = cacheEntry{result: result, expires: time.Now().Add(s.Config.CacheTTL)}
	s.cacheMu.Unlock()
}

func cacheKey(req Request, factsJSON []byte) string {
	h := sha1.Sum(append([]byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", req.TargetType, req.TargetID, req.Namespace, req.Surface, req.SegmentID, req.From.Format(time.RFC3339), req.To.Format(time.RFC3339))), factsJSON...))
	return hex.EncodeToString(h[:])
}

func (s *Service) selectModel(factsLen int) string {
	if s.Config.ModelPrimary == "" {
		return "fallback"
	}
	if factsLen > 2000 && s.Config.ModelEscalate != "" {
		return s.Config.ModelEscalate
	}
	return s.Config.ModelPrimary
}

const defaultSystemPrompt = "You are a reliability analyst for our Recommendation System. Use ONLY the provided FACTS. Do NOT invent data. When a plausible cause lacks evidence, mark it uncertain. Be concise, action-oriented, and cite evidence_id for claims. Output markdown with sections: Summary, Status, Key findings, Likely causes, Suggested fix (only if warranted). Never include private reasoning."

func renderUserPrompt(req Request, factsJSON string) string {
	segment := req.SegmentID
	if segment == "" {
		segment = "(all)"
	}
	return fmt.Sprintf("Investigate why %s \"%s\" appears to be “not working” on surface \"%s\" in namespace \"%s\" during %s..%s for segment \"%s\".\n\nFACTS (compact JSON):\n%s\n\nIf facts are insufficient, say so explicitly and suggest the top 3 diagnostics.\n", strings.ToLower(req.TargetType), req.TargetID, req.Surface, req.Namespace, req.From.Format(time.RFC3339), req.To.Format(time.RFC3339), segment, factsJSON)
}

func fallbackMarkdown(facts FactsPack, evidence []Evidence) string {
	status := "unknown"
	if facts.Metrics.Impressions == 0 {
		status = "not_working"
	} else if facts.Metrics.Clicks == 0 {
		status = "degraded"
	} else {
		status = "working_as_configured"
	}

	summary := fmt.Sprintf("Observed %d impressions and %d clicks for %s on %s/%s during %s–%s.", facts.Metrics.Impressions, facts.Metrics.Clicks, facts.Target.ID, facts.Context.Namespace, facts.Context.Surface, facts.Window.From, facts.Window.To)

	var keyFindings []string
	for _, ev := range evidence {
		keyFindings = append(keyFindings, fmt.Sprintf("- %s · evidence_id=[%s]", ev.Message, ev.EvidenceID))
		if len(keyFindings) >= 5 {
			break
		}
	}
	for len(keyFindings) < 3 {
		keyFindings = append(keyFindings, "- No additional findings · evidence_id=[none]")
	}

	var likelyCauses []string
	for _, ev := range evidence {
		likelyCauses = append(likelyCauses, fmt.Sprintf("- %s · confidence=0.4 · evidence_id=[%s]", ev.Message, ev.EvidenceID))
		if len(likelyCauses) >= 3 {
			break
		}
	}
	for len(likelyCauses) < 2 {
		likelyCauses = append(likelyCauses, "- Cause unclear · confidence=0.1 · evidence_id=[none]")
	}

	suggested := "None warranted."
	if facts.Metrics.Impressions == 0 {
		suggested = "Investigate upstream eligibility filters or rules blocking delivery."
	}

	builder := strings.Builder{}
	builder.WriteString("## Summary\n\n")
	builder.WriteString(summary)
	builder.WriteString("\n\n## Status\n\n")
	builder.WriteString(status)
	builder.WriteString("\n\n## Key findings\n\n")
	for _, entry := range keyFindings {
		builder.WriteString(entry)
		builder.WriteString("\n")
	}
	builder.WriteString("\n## Likely causes\n\n")
	for _, entry := range likelyCauses {
		builder.WriteString(entry)
		builder.WriteString("\n")
	}
	builder.WriteString("\n## Suggested fix\n\n")
	builder.WriteString("- ")
	builder.WriteString(suggested)
	builder.WriteString("\n")
	return builder.String()
}

// NullClient is a fallback LLM client that always returns an error.
type NullClient struct{}

// Generate implements LLMClient.
func (NullClient) Generate(context.Context, string, string, string, int) (string, error) {
	return "", errors.New("llm client not configured")
}
