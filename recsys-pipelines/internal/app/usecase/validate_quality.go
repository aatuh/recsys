package usecase

import (
	"context"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/validator"
)

type ValidateQuality struct {
	rt        runtime.Runtime
	validator validator.Validator
}

func NewValidateQuality(rt runtime.Runtime, v validator.Validator) *ValidateQuality {
	return &ValidateQuality{rt: rt, validator: v}
}

func (uc *ValidateQuality) Execute(ctx context.Context, tenant, surface string, w windows.Window) error {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "validate: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
		logger.Field{Key: "start", Value: w.Start.Format(time.RFC3339)},
		logger.Field{Key: "end", Value: w.End.Format(time.RFC3339)},
	)
	if err := uc.validator.ValidateCanonical(ctx, tenant, surface, w); err != nil {
		return err
	}
	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "validate: done", logger.Field{Key: "duration_ms", Value: dur.Milliseconds()})
	return nil
}
