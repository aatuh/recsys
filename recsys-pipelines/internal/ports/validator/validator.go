package validator

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type Validator interface {
	ValidateCanonical(ctx context.Context, tenant, surface string, w windows.Window) error
	ValidateArtifact(ctx context.Context, ref artifacts.Ref, artifactJSON []byte) error
}
