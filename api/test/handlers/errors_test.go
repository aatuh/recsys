package handlers

import (
	"net/http"
	"strings"
	"testing"

	"recsys/specs/endpoints"
	"recsys/specs/types"
	"recsys/test/shared"
)

func TestEventTypes_Validation_CheckViolation(t *testing.T) {
	client := shared.NewTestClient(t)

	// Negative weight should violate check constraint -> 422
	client.DoRequestWithStatus(t, http.MethodPost, endpoints.EventTypesUpsert, types.EventTypeConfigUpsertRequest{
		Namespace: "default",
		Types: []types.EventTypeConfig{
			{Type: 0, Weight: -0.5},
		},
	}, http.StatusUnprocessableEntity)
}

func TestItemsUpsert_MalformedJSON_400(t *testing.T) {
	client := shared.NewTestClient(t)

	malformedJSON := `{"namespace": "default", "items": [ {"item_id": 123 } ]` // missing closing ]
	client.DoRawRequest(t, http.MethodPost, endpoints.ItemsUpsert, strings.NewReader(malformedJSON), http.StatusBadRequest)
}
