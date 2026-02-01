package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/cache"
	"github.com/aatuh/recsys-suite/api/internal/objectstore"
)

type LoaderConfig struct {
	ManifestTemplate string
	ManifestTTL      time.Duration
	ArtifactTTL      time.Duration
	MaxBytes         int
}

type Loader struct {
	reader           objectstore.Reader
	manifestTemplate string
	manifestCache    *cache.TTLCache[manifestKey, ManifestV1]
	popCache         *cache.TTLCache[string, PopularityArtifactV1]
	coocCache        *cache.TTLCache[string, CoocArtifactV1]
	implicitCache    *cache.TTLCache[string, ImplicitArtifactV1]
	contentCache     *cache.TTLCache[string, ContentArtifactV1]
	sessionCache     *cache.TTLCache[string, SessionSeqArtifactV1]
	manifestTTL      time.Duration
	artifactTTL      time.Duration
	maxBytes         int
}

type manifestKey struct {
	Tenant  string
	Surface string
}

func NewLoader(reader objectstore.Reader, cfg LoaderConfig) *Loader {
	return &Loader{
		reader:           reader,
		manifestTemplate: strings.TrimSpace(cfg.ManifestTemplate),
		manifestCache:    cache.NewTTL[manifestKey, ManifestV1](nil),
		popCache:         cache.NewTTL[string, PopularityArtifactV1](nil),
		coocCache:        cache.NewTTL[string, CoocArtifactV1](nil),
		implicitCache:    cache.NewTTL[string, ImplicitArtifactV1](nil),
		contentCache:     cache.NewTTL[string, ContentArtifactV1](nil),
		sessionCache:     cache.NewTTL[string, SessionSeqArtifactV1](nil),
		manifestTTL:      cfg.ManifestTTL,
		artifactTTL:      cfg.ArtifactTTL,
		maxBytes:         cfg.MaxBytes,
	}
}

func (l *Loader) Invalidate(tenant, surface string) int {
	if l == nil {
		return 0
	}
	n := 0
	if l.manifestCache != nil {
		n += l.manifestCache.Invalidate(func(k manifestKey, _ ManifestV1) bool {
			if tenant != "" && k.Tenant != tenant {
				return false
			}
			if surface != "" && k.Surface != surface {
				return false
			}
			return true
		})
	}
	if l.popCache != nil {
		n += l.popCache.Invalidate(func(key string, _ PopularityArtifactV1) bool { return true })
	}
	if l.coocCache != nil {
		n += l.coocCache.Invalidate(func(key string, _ CoocArtifactV1) bool { return true })
	}
	if l.implicitCache != nil {
		n += l.implicitCache.Invalidate(func(key string, _ ImplicitArtifactV1) bool { return true })
	}
	if l.contentCache != nil {
		n += l.contentCache.Invalidate(func(key string, _ ContentArtifactV1) bool { return true })
	}
	if l.sessionCache != nil {
		n += l.sessionCache.Invalidate(func(key string, _ SessionSeqArtifactV1) bool { return true })
	}
	return n
}

func (l *Loader) ManifestURI(tenant, surface string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("loader is nil")
	}
	if l.manifestTemplate == "" {
		return "", fmt.Errorf("manifest template is required")
	}
	out := l.manifestTemplate
	out = strings.ReplaceAll(out, "{tenant}", tenant)
	out = strings.ReplaceAll(out, "{surface}", surface)
	return out, nil
}

func (l *Loader) LoadManifest(ctx context.Context, tenant, surface string) (ManifestV1, bool, error) {
	if l == nil {
		return ManifestV1{}, false, fmt.Errorf("loader is nil")
	}
	key := manifestKey{Tenant: tenant, Surface: surface}
	if val, ok := l.manifestCache.Get(key); ok {
		return val, true, nil
	}
	uri, err := l.ManifestURI(tenant, surface)
	if err != nil {
		return ManifestV1{}, false, err
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return ManifestV1{}, false, nil
		}
		return ManifestV1{}, false, err
	}
	var manifest ManifestV1
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ManifestV1{}, false, wrapManifestError(err)
	}
	if err := manifest.Validate(); err != nil {
		return ManifestV1{}, false, wrapManifestError(err)
	}
	if manifest.Tenant != tenant || manifest.Surface != surface {
		return ManifestV1{}, false, wrapManifestError(fmt.Errorf("manifest tenant/surface mismatch"))
	}
	if l.manifestTTL > 0 {
		l.manifestCache.Set(key, manifest, l.manifestTTL)
	}
	return manifest, true, nil
}

func (l *Loader) LoadPopularity(ctx context.Context, uri string) (PopularityArtifactV1, bool, error) {
	if l == nil {
		return PopularityArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return PopularityArtifactV1{}, false, nil
	}
	if val, ok := l.popCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return PopularityArtifactV1{}, false, nil
		}
		return PopularityArtifactV1{}, false, err
	}
	var art PopularityArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return PopularityArtifactV1{}, false, wrapArtifactError(err)
	}
	if err := art.Validate(); err != nil {
		return PopularityArtifactV1{}, false, wrapArtifactError(err)
	}
	if l.artifactTTL > 0 {
		l.popCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}

func (l *Loader) LoadCooc(ctx context.Context, uri string) (CoocArtifactV1, bool, error) {
	if l == nil {
		return CoocArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return CoocArtifactV1{}, false, nil
	}
	if val, ok := l.coocCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return CoocArtifactV1{}, false, nil
		}
		return CoocArtifactV1{}, false, err
	}
	var art CoocArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return CoocArtifactV1{}, false, wrapArtifactError(err)
	}
	if err := art.Validate(); err != nil {
		return CoocArtifactV1{}, false, wrapArtifactError(err)
	}
	if l.artifactTTL > 0 {
		l.coocCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}

func (l *Loader) LoadImplicit(ctx context.Context, uri string) (ImplicitArtifactV1, bool, error) {
	if l == nil {
		return ImplicitArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return ImplicitArtifactV1{}, false, nil
	}
	if val, ok := l.implicitCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return ImplicitArtifactV1{}, false, nil
		}
		return ImplicitArtifactV1{}, false, err
	}
	var art ImplicitArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return ImplicitArtifactV1{}, false, wrapArtifactError(err)
	}
	if err := art.Validate(); err != nil {
		return ImplicitArtifactV1{}, false, wrapArtifactError(err)
	}
	if l.artifactTTL > 0 {
		l.implicitCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}

func (l *Loader) LoadContent(ctx context.Context, uri string) (ContentArtifactV1, bool, error) {
	if l == nil {
		return ContentArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return ContentArtifactV1{}, false, nil
	}
	if val, ok := l.contentCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return ContentArtifactV1{}, false, nil
		}
		return ContentArtifactV1{}, false, err
	}
	var art ContentArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return ContentArtifactV1{}, false, wrapArtifactError(err)
	}
	if err := art.Validate(); err != nil {
		return ContentArtifactV1{}, false, wrapArtifactError(err)
	}
	if l.artifactTTL > 0 {
		l.contentCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}

func (l *Loader) LoadSessionSeq(ctx context.Context, uri string) (SessionSeqArtifactV1, bool, error) {
	if l == nil {
		return SessionSeqArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return SessionSeqArtifactV1{}, false, nil
	}
	if val, ok := l.sessionCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return SessionSeqArtifactV1{}, false, nil
		}
		return SessionSeqArtifactV1{}, false, err
	}
	var art SessionSeqArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return SessionSeqArtifactV1{}, false, wrapArtifactError(err)
	}
	if err := art.Validate(); err != nil {
		return SessionSeqArtifactV1{}, false, wrapArtifactError(err)
	}
	if l.artifactTTL > 0 {
		l.sessionCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}
