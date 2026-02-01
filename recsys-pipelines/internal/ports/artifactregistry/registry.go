package artifactregistry

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
)

type Registry interface {
	Record(ctx context.Context, ref artifacts.Ref) error
	LoadManifest(ctx context.Context, tenant, surface string) (artifacts.ManifestV1, bool, error)
	SwapManifest(ctx context.Context, tenant, surface string, next artifacts.ManifestV1) error
}
