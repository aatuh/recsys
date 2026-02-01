package noop

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/validator"
)

type NoopValidator struct{}

var _ validator.Validator = NoopValidator{}

func (NoopValidator) ValidateCanonical(ctx context.Context, _ string, _ string, _ windows.Window) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (NoopValidator) ValidateArtifact(ctx context.Context, _ artifacts.Ref, _ []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
