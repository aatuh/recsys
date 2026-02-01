package usecase

import (
	"context"
	"path"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/artifactregistry"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/objectstore"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/validator"
)

type PublishArtifacts struct {
	rt        runtime.Runtime
	store     objectstore.ObjectStore
	registry  artifactregistry.Registry
	validator validator.Validator
}

func NewPublishArtifacts(
	rt runtime.Runtime,
	store objectstore.ObjectStore,
	registry artifactregistry.Registry,
	validator validator.Validator,
) *PublishArtifacts {
	return &PublishArtifacts{rt: rt, store: store, registry: registry, validator: validator}
}

type ArtifactBlob struct {
	Ref  artifacts.Ref
	JSON []byte
}

type PublishInput struct {
	Tenant  string
	Surface string

	Popularity *ArtifactBlob
	Cooc       *ArtifactBlob
	Implicit   *ArtifactBlob
	Content    *ArtifactBlob
	SessionSeq *ArtifactBlob
}

func (uc *PublishArtifacts) Execute(ctx context.Context, in PublishInput) error {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "publish: start",
		logger.Field{Key: "tenant", Value: in.Tenant},
		logger.Field{Key: "surface", Value: in.Surface},
	)

	current := map[string]string{}
	existing, ok, err := uc.registry.LoadManifest(ctx, in.Tenant, in.Surface)
	if err != nil {
		return err
	}
	if ok {
		for k, v := range existing.Current {
			current[k] = v
		}
	}

	publishOne := func(kind string, blob *ArtifactBlob) (string, error) {
		if blob == nil {
			return "", nil
		}
		key := path.Join(in.Tenant, in.Surface, kind, blob.Ref.Version+".json")
		uri, err := uc.store.Put(ctx, key, "application/json", blob.JSON)
		if err != nil {
			return "", err
		}
		blob.Ref.URI = uri

		if err := uc.validator.ValidateArtifact(ctx, blob.Ref, blob.JSON); err != nil {
			return "", err
		}
		if err := uc.registry.Record(ctx, blob.Ref); err != nil {
			return "", err
		}
		return uri, nil
	}

	if in.Popularity != nil {
		uri, err := publishOne("popularity", in.Popularity)
		if err != nil {
			return err
		}
		current["popularity"] = uri
	}
	if in.Cooc != nil {
		uri, err := publishOne("cooc", in.Cooc)
		if err != nil {
			return err
		}
		current["cooc"] = uri
	}
	if in.Implicit != nil {
		uri, err := publishOne("implicit", in.Implicit)
		if err != nil {
			return err
		}
		current["implicit"] = uri
	}
	if in.Content != nil {
		uri, err := publishOne("content_sim", in.Content)
		if err != nil {
			return err
		}
		current["content_sim"] = uri
	}
	if in.SessionSeq != nil {
		uri, err := publishOne("session_seq", in.SessionSeq)
		if err != nil {
			return err
		}
		current["session_seq"] = uri
	}

	next := artifacts.NewManifest(in.Tenant, in.Surface, current, uc.rt.Clock.NowUTC())
	if err := uc.registry.SwapManifest(ctx, in.Tenant, in.Surface, next); err != nil {
		return err
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "publish: done", logger.Field{Key: "duration_ms", Value: dur.Milliseconds()})
	return nil
}
