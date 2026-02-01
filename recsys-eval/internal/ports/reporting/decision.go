package reporting

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/decision"
)

// DecisionWriter persists decision artifacts.
type DecisionWriter interface {
	Write(ctx context.Context, artifact decision.Artifact, path string) error
}
