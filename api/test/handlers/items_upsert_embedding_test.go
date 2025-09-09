package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"recsys/internal/http/handlers"
)

// Test that ItemsUpsert rejects embeddings with the wrong dimension.
// This returns before touching the store, so we do not need a DB.
func TestItemsUpsert_EmbeddingDimMismatch(t *testing.T) {
	h := &handlers.Handler{} // Store not needed for early-return branch.

	// Build valid JSON with a 3-dim embedding to trigger the check.
	payload := map[string]any{
		"namespace": "default",
		"items": []map[string]any{
			{
				"item_id":   "i_bad",
				"available": true,
				"embedding": []float64{1.0, 0.0, 0.0}, // 3 instead of 384
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/items:upsert",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ItemsUpsert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s",
			rec.Code, rec.Body.String())
	}

	// Parse response and accept either "error" or "id" field names.
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not JSON: %v; body=%s",
			err, rec.Body.String())
	}

	var code string
	if v, ok := resp["error"].(string); ok {
		code = v
	} else if v, ok := resp["id"].(string); ok {
		code = v
	} else {
		t.Fatalf("missing error code; body=%s", rec.Body.String())
	}

	if code != "embedding_dim_mismatch" {
		t.Fatalf("want error=embedding_dim_mismatch, got %q; body=%s",
			code, rec.Body.String())
	}
}
