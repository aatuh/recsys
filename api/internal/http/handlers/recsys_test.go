package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aatuh/recsys-suite/api/internal/auth"
)

type noopValidator struct{}

func (noopValidator) Validate(context.Context, interface{}) error              { return nil }
func (noopValidator) ValidateStruct(context.Context, interface{}) error        { return nil }
func (noopValidator) ValidateField(context.Context, interface{}, string) error { return nil }

func TestRecommendRejectsFullExplainWithoutAdminRole(t *testing.T) {
	h := NewRecsysHandler(nil, nil, noopValidator{}, WithExplainControls(50, true, "admin"))
	body := []byte(`{"surface":"home","user":{"user_id":"user-1"},"options":{"explain":"full"}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/recommend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithInfo(req.Context(), auth.Info{Roles: []string{"viewer"}}))
	rec := httptest.NewRecorder()

	h.recommend(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "RECSYS_FORBIDDEN") {
		t.Fatalf("body = %s, want forbidden problem code", rec.Body.String())
	}
}
