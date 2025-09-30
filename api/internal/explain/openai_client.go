package explain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// OpenAIClient calls the OpenAI Responses API.
type OpenAIClient struct {
	HTTP    *http.Client
	APIKey  string
	BaseURL string
	Logger  *zap.Logger
}

type openAIRequest struct {
	Model           string        `json:"model"`
	Input           []openAIInput `json:"input"`
	MaxOutputTokens int           `json:"max_output_tokens,omitempty"`
}

type openAIInput struct {
	Role    string          `json:"role"`
	Content []openAIContent `json:"content"`
}

type openAIContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIResponse struct {
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// Generate implements LLMClient.
func (c *OpenAIClient) Generate(ctx context.Context, model string, systemPrompt string, userPrompt string, maxTokens int) (string, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return "", fmt.Errorf("openai api key missing")
	}

	baseURL := c.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1/responses"
	}

	reqBody := openAIRequest{
		Model: model,
		Input: []openAIInput{
			{
				Role:    "system",
				Content: []openAIContent{{Type: "input_text", Text: systemPrompt}},
			},
			{
				Role:    "user",
				Content: []openAIContent{{Type: "input_text", Text: userPrompt}},
			},
		},
	}
	if maxTokens > 0 {
		reqBody.MaxOutputTokens = maxTokens
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 300 {
		if c.Logger != nil {
			c.Logger.Warn("openai request failed", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		}
		return "", fmt.Errorf("openai error: %s", resp.Status)
	}

	var parsed openAIResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("openai error: %s", parsed.Error.Message)
	}

	for _, chunk := range parsed.Output {
		for _, content := range chunk.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text, nil
			}
		}
	}

	return "", fmt.Errorf("openai: empty response")
}
