package runtime

import (
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/clock"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/metrics"
)

type Runtime struct {
	Clock   clock.Clock
	Logger  logger.Logger
	Metrics metrics.Metrics
}
