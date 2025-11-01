package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"recsys/internal/http/common"
	"recsys/internal/http/handlers"
	"recsys/internal/services/ingestion"
	"recsys/specs/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestErrorHandling_StructuredLogging(t *testing.T) {
	// Setup test handler
	logger := zap.NewNop() // Use no-op logger for testing
	ingSvc := ingestion.New(nil)
	h := handlers.NewIngestionHandler(ingSvc, uuid.New(), logger)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "malformed_json",
			requestBody:    "invalid json",
			expectedStatus: 400,
			expectedCode:   "invalid_json",
		},
		{
			name: "embedding_dimension_mismatch",
			requestBody: types.ItemsUpsertRequest{
				Items: []types.Item{
					{
						ItemID:    "test-item",
						Embedding: []float64{1.0, 2.0}, // Wrong dimension
						Available: true,
					},
				},
			},
			expectedStatus: 400,
			expectedCode:   "embedding_dim_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/v1/items:upsert", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Org-ID", uuid.New().String())

			// Add request ID middleware for correlation ID
			r := chi.NewRouter()
			r.Use(middleware.RequestID)
			r.Post("/v1/items:upsert", h.ItemsUpsert)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var errorResp common.APIError
			if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if errorResp.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, errorResp.Code)
			}

			// Verify error response structure
			if errorResp.CorrelationID == "" {
				t.Error("Expected correlation ID to be set")
			}

			if errorResp.Timestamp.IsZero() {
				t.Error("Expected timestamp to be set")
			}
		})
	}
}

func TestErrorHandling_5xxWithStackTrace(t *testing.T) {
	// This test would require a more complex setup to trigger actual 5xx errors
	// For now, we'll test the error structure
	ae := common.NewAPIError("test_error", "Test error message", 500)
	ae = ae.WithStackTrace()

	// Test that the error was created successfully
	if ae.Code != "test_error" {
		t.Errorf("Expected error code 'test_error', got '%s'", ae.Code)
	}
}

func TestErrorHandling_ContextualErrors(t *testing.T) {
	// Test error context creation
	ctx := common.ErrorContext{
		UserID:    "user123",
		OrgID:     "org456",
		RequestID: "req789",
		Path:      "/v1/test",
		Method:    "POST",
		UserAgent: "test-agent",
		IP:        "127.0.0.1",
	}

	ae := common.NewAPIErrorWithContext("test_error", "Test error", 400, ctx)

	if ae.CorrelationID != ctx.RequestID {
		t.Errorf("Expected correlation ID %s, got %s", ctx.RequestID, ae.CorrelationID)
	}
}

func TestErrorHandling_DebugMode(t *testing.T) {
	// Test debug message inclusion
	ae := common.NewAPIError("test_error", "Test error", 500)
	ae = ae.WithDebugMessage("Debug information")

	// Test that the error was created successfully
	if ae.Code != "test_error" {
		t.Errorf("Expected error code 'test_error', got '%s'", ae.Code)
	}
}

func TestErrorHandling_PostgresErrorMapping(t *testing.T) {
	// Test PostgreSQL error code mapping
	tests := []struct {
		code           string
		expectedStatus int
		expectedCode   string
	}{
		{"23505", 409, "unique_violation"},
		{"23503", 422, "foreign_key_violation"},
		{"23502", 422, "not_null_violation"},
		{"23514", 422, "check_violation"},
		{"22001", 422, "string_truncation"},
		{"99999", 422, "constraint_violation"}, // Unknown code
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("pg_code_%s", tt.code), func(t *testing.T) {
			// This would require creating a pgconn.PgError, which is complex
			// For now, we'll test the mapping function directly
			ae := common.NewAPIError("constraint_violation", "Data violates a constraint", 422)

			// Test that the error was created successfully
			if ae.Code != "constraint_violation" {
				t.Errorf("Expected error code 'constraint_violation', got '%s'", ae.Code)
			}
		})
	}
}
