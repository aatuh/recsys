package workflow

import (
	"context"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/staging"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type Pipeline struct {
	RT           runtime.Runtime
	ArtifactsDir string

	Ingest   *usecase.IngestEvents
	Validate *usecase.ValidateQuality
	Pop      *usecase.ComputePopularity
	Cooc     *usecase.ComputeCooc
	Signals  *usecase.PersistSignals
	Publish  *usecase.PublishArtifacts
}

func (p *Pipeline) RunDay(ctx context.Context, tenant, surface, segment string, w windows.Window) error {
	if err := w.Validate(); err != nil {
		return err
	}
	if err := p.Ingest.Execute(ctx, tenant, surface, w); err != nil {
		return err
	}
	if err := p.Validate.Execute(ctx, tenant, surface, w); err != nil {
		return err
	}

	popRef, popJSON, err := p.Pop.Execute(ctx, tenant, surface, segment, w)
	if err != nil {
		return err
	}
	coocRef, coocJSON, err := p.Cooc.Execute(ctx, tenant, surface, segment, w)
	if err != nil {
		return err
	}

	if p.Signals != nil {
		if err := p.Signals.Execute(ctx, popJSON, coocJSON); err != nil {
			return err
		}
	}

	if p.ArtifactsDir != "" {
		st := staging.New(p.ArtifactsDir)
		if _, err := st.Put(ctx, popRef, popJSON); err != nil {
			return err
		}
		if _, err := st.Put(ctx, coocRef, coocJSON); err != nil {
			return err
		}
	}

	in := usecase.PublishInput{
		Tenant:     tenant,
		Surface:    surface,
		Popularity: &usecase.ArtifactBlob{Ref: popRef, JSON: popJSON},
		Cooc:       &usecase.ArtifactBlob{Ref: coocRef, JSON: coocJSON},
	}
	return p.Publish.Execute(ctx, in)
}
