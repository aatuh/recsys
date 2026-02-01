package validation

import (
	"fmt"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/api/src/specs/types"
)

func TestNormalizeRecommendDefaults(t *testing.T) {
	req := &types.RecommendRequest{
		Surface: "home",
		User: &types.UserRef{
			UserID: randID("user"),
		},
	}

	out, warnings, err := NormalizeRecommendRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Segment != "default" {
		t.Fatalf("expected default segment, got %q", out.Segment)
	}
	if out.K != 20 {
		t.Fatalf("expected default k=20, got %d", out.K)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warnings for defaults")
	}
}

func TestNormalizeRecommendRequiresUser(t *testing.T) {
	req := &types.RecommendRequest{Surface: "home"}
	_, _, err := NormalizeRecommendRequest(req)
	if err == nil {
		t.Fatalf("expected error")
	}
	verr, ok := err.(Error)
	if !ok {
		t.Fatalf("expected validation.Error, got %T", err)
	}
	if verr.Status != 422 {
		t.Fatalf("expected status 422, got %d", verr.Status)
	}
}

func TestNormalizeRecommendContextNow(t *testing.T) {
	req := &types.RecommendRequest{
		Surface: "home",
		User:    &types.UserRef{UserID: randID("user")},
		Context: &types.RequestContext{Now: "not-a-time"},
	}
	_, _, err := NormalizeRecommendRequest(req)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func randID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func TestNormalizeSimilarRequiresItem(t *testing.T) {
	req := &types.SimilarRequest{Surface: "pdp"}
	_, _, err := NormalizeSimilarRequest(req)
	if err == nil {
		t.Fatalf("expected error")
	}
}
