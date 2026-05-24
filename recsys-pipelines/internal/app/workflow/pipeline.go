package workflow

import (
	"context"
	"fmt"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
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
	Implicit *usecase.ComputeImplicit
	Content  *usecase.ComputeContentSim
	Session  *usecase.ComputeSessionSeq
	Signals  *usecase.PersistSignals
	Publish  *usecase.PublishArtifacts

	Artifacts config.ArtifactSelection
}

func (p *Pipeline) RunDay(ctx context.Context, tenant, surface, segment string, w windows.Window) error {
	if err := w.Validate(); err != nil {
		return err
	}
	selection := p.Artifacts
	if !selection.Popularity && !selection.Cooc && !selection.Implicit && !selection.ContentSim && !selection.SessionSeq {
		selection.Popularity = true
		selection.Cooc = true
	}
	if err := p.Ingest.Execute(ctx, tenant, surface, w); err != nil {
		return err
	}
	if err := p.Validate.Execute(ctx, tenant, surface, w); err != nil {
		return err
	}

	var popBlob, coocBlob, implicitBlob, contentBlob, sessionBlob *usecase.ArtifactBlob
	if selection.Popularity {
		if p.Pop == nil {
			return fmt.Errorf("popularity artifact enabled but compute use case is not configured")
		}
		popRef, popJSON, err := p.Pop.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		popBlob = &usecase.ArtifactBlob{Ref: popRef, JSON: popJSON}
	}
	if selection.Cooc {
		if p.Cooc == nil {
			return fmt.Errorf("cooc artifact enabled but compute use case is not configured")
		}
		coocRef, coocJSON, err := p.Cooc.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		coocBlob = &usecase.ArtifactBlob{Ref: coocRef, JSON: coocJSON}
	}
	if selection.Implicit {
		if p.Implicit == nil {
			return fmt.Errorf("implicit artifact enabled but compute use case is not configured")
		}
		implicitRef, implicitJSON, err := p.Implicit.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		implicitBlob = &usecase.ArtifactBlob{Ref: implicitRef, JSON: implicitJSON}
	}
	if selection.ContentSim {
		if p.Content == nil {
			return fmt.Errorf("content_sim artifact enabled but compute use case is not configured")
		}
		contentRef, contentJSON, err := p.Content.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		contentBlob = &usecase.ArtifactBlob{Ref: contentRef, JSON: contentJSON}
	}
	if selection.SessionSeq {
		if p.Session == nil {
			return fmt.Errorf("session_seq artifact enabled but compute use case is not configured")
		}
		sessionRef, sessionJSON, err := p.Session.Execute(ctx, tenant, surface, segment, w)
		if err != nil {
			return err
		}
		sessionBlob = &usecase.ArtifactBlob{Ref: sessionRef, JSON: sessionJSON}
	}

	if p.Signals != nil && popBlob != nil && coocBlob != nil {
		if err := p.Signals.Execute(ctx, popBlob.JSON, coocBlob.JSON); err != nil {
			return err
		}
	}

	if p.ArtifactsDir != "" {
		st := staging.New(p.ArtifactsDir)
		for _, blob := range []*usecase.ArtifactBlob{popBlob, coocBlob, implicitBlob, contentBlob, sessionBlob} {
			if blob == nil {
				continue
			}
			if _, err := st.Put(ctx, blob.Ref, blob.JSON); err != nil {
				return err
			}
		}
	}

	in := usecase.PublishInput{
		Tenant:     tenant,
		Surface:    surface,
		Popularity: popBlob,
		Cooc:       coocBlob,
		Implicit:   implicitBlob,
		Content:    contentBlob,
		SessionSeq: sessionBlob,
	}
	return p.Publish.Execute(ctx, in)
}
