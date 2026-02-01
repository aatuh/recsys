package json

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/decision"
)

// DecisionWriter writes decision artifacts as JSON.
type DecisionWriter struct{}

func (DecisionWriter) Write(_ context.Context, artifact decision.Artifact, path string) error {
	// #nosec G304 -- output path provided by operator
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(artifact)
}
