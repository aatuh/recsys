package json

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

// Writer writes reports as JSON.
type Writer struct{}

func (Writer) Write(_ context.Context, rep report.Report, path string) error {
	// #nosec G304 -- output path provided by operator
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}
