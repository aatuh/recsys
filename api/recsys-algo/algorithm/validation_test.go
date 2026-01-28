package algorithm

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestRequestValidateReturnsValidationErrors(t *testing.T) {
	req := Request{
		OrgID:              uuid.Nil,
		K:                  -1,
		StarterBlendWeight: 2,
		RecentEventCount:   -5,
		Blend:              &BlendWeights{Pop: -1, Cooc: 0, Similarity: 0},
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
	if len(verrs) < 3 {
		t.Fatalf("expected multiple validation errors, got %#v", verrs)
	}
}

func TestConfigValidateReturnsValidationErrors(t *testing.T) {
	cfg := Config{
		BlendAlpha:                 -1,
		MMRLambda:                  2,
		ProfileColdStartMultiplier: 1.5,
		MaxK:                       -2,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
	if len(verrs) < 3 {
		t.Fatalf("expected multiple validation errors, got %#v", verrs)
	}
}
