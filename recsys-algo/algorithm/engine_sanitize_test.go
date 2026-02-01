package algorithm

import (
	"testing"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
)

func TestSanitizeRequestAppliesCaps(t *testing.T) {
	eng := NewEngine(Config{
		MaxK:               3,
		MaxExcludeIDs:      2,
		MaxAnchorsInjected: 2,
	}, nil, nil)

	req := Request{
		K:             10,
		InjectAnchors: true,
		AnchorItemIDs: []string{" a ", "b", "a", "", "c"},
		Constraints: &recmodel.PopConstraints{
			ExcludeItemIDs: []string{"x", "y", "z"},
		},
	}

	sanitized := eng.sanitizeRequest(req)
	if sanitized.K != 3 {
		t.Fatalf("expected K clamped to 3, got %d", sanitized.K)
	}
	if sanitized.Constraints == nil || len(sanitized.Constraints.ExcludeItemIDs) != 2 {
		t.Fatalf("expected 2 exclude IDs, got %#v", sanitized.Constraints)
	}
	if len(sanitized.AnchorItemIDs) != 2 {
		t.Fatalf("expected 2 anchors, got %#v", sanitized.AnchorItemIDs)
	}
	if sanitized.AnchorItemIDs[0] != "a" || sanitized.AnchorItemIDs[1] != "b" {
		t.Fatalf("unexpected anchors %#v", sanitized.AnchorItemIDs)
	}
}
