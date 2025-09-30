package explain

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: roundTripFunc(fn), Timeout: time.Second}
}

func TestOpenAIClientGenerate(t *testing.T) {
	client := &OpenAIClient{
		HTTP: newTestClient(func(r *http.Request) (*http.Response, error) {
			require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			body := io.NopCloser(strings.NewReader(`{"output":[{"content":[{"type":"output_text","text":"hello"}]}]}`))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
				Header:     make(http.Header),
			}, nil
		}),
		APIKey:  "test-key",
		BaseURL: "https://example.com",
		Logger:  zap.NewNop(),
	}

	text, err := client.Generate(context.Background(), "o4-mini", "system", "user", 100)
	require.NoError(t, err)
	require.Equal(t, "hello", text)
}

func TestOpenAIClientHTTPError(t *testing.T) {
	client := &OpenAIClient{
		HTTP: newTestClient(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"error":"bad"}`)),
				Header:     make(http.Header),
			}, nil
		}),
		APIKey:  "key",
		BaseURL: "https://example.com",
	}

	_, err := client.Generate(context.Background(), "o4-mini", "system", "user", 50)
	require.Error(t, err)
}
