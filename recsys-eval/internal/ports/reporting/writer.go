package reporting

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
)

// Writer persists evaluation reports.
type Writer interface {
	Write(ctx context.Context, report report.Report, path string) error
}
