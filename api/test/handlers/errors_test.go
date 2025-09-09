package handlers

import (
	"net/http"
	"strings"
	"testing"

	"recsys/test/shared"
)

func TestEventTypes_Validation_CheckViolation(t *testing.T) {
	client := shared.NewTestClient(t)

	// Negative weight should violate check constraint -> 422
	client.DoRequestWithStatus(t, http.MethodPost, "/v1/event-types:upsert", map[string]any{
		"namespace": "default",
		"types": []map[string]any{
			{"type": 0, "weight": -0.5},
		},
	}, http.StatusUnprocessableEntity)
}

func TestItemsUpsert_MalformedJSON_400(t *testing.T) {
	client := shared.NewTestClient(t)

	malformedJSON := `{"namespace": "default", "items": [ {"item_id": 123 } ]` // missing closing ]
	client.DoRawRequest(t, http.MethodPost, "/v1/items:upsert", strings.NewReader(malformedJSON), http.StatusBadRequest)
}
