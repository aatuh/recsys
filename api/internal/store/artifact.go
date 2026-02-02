package store

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/artifacts"

	"github.com/aatuh/api-toolkit/v2/authorization"
	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
	"github.com/google/uuid"
)

// ArtifactAlgoStore provides recsys-algo stores backed by artifact manifests.
type ArtifactAlgoStore struct {
	loader   *artifacts.Loader
	tagStore recmodel.TagStore
}

// NewArtifactAlgoStore constructs an artifact-backed store with a tag store fallback.
func NewArtifactAlgoStore(loader *artifacts.Loader, tagStore recmodel.TagStore) *ArtifactAlgoStore {
	return &ArtifactAlgoStore{loader: loader, tagStore: tagStore}
}

// PopularityTopK returns candidates ordered by popularity counts in the artifact.
func (s *ArtifactAlgoStore) PopularityTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	halfLifeDays float64,
	k int,
	c *recmodel.PopConstraints,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		return nil, nil
	}
	if s == nil || s.loader == nil {
		return nil, nil
	}
	_ = halfLifeDays
	tenant := tenantKey(ctx, orgID)
	if tenant == "" {
		return nil, nil
	}
	ns = normalizeNamespace(ns)

	items, _, err := s.loadPopularity(ctx, tenant, ns)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 && ns != "default" {
		fallbackNS := "default"
		fallbackItems, _, err := s.loadPopularity(ctx, tenant, fallbackNS)
		if err != nil {
			return nil, err
		}
		if len(fallbackItems) > 0 {
			ns = fallbackNS
			items = fallbackItems
		}
	}
	if len(items) == 0 {
		return nil, nil
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Count > items[j].Count
	})

	return s.filterPopularity(ctx, orgID, ns, k, items, c)
}

// ListItemsTags delegates to the configured tag store.
func (s *ArtifactAlgoStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	if s == nil || s.tagStore == nil {
		return map[string]recmodel.ItemTags{}, nil
	}
	return s.tagStore.ListItemsTags(ctx, orgID, ns, itemIDs)
}

// CooccurrenceTopKWithin returns similar items using co-visitation artifacts.
func (s *ArtifactAlgoStore) CooccurrenceTopKWithin(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	anchor string,
	k int,
	since time.Time,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		return nil, nil
	}
	anchor = strings.TrimSpace(anchor)
	if anchor == "" {
		return nil, nil
	}
	if s == nil || s.loader == nil {
		return nil, recmodel.ErrFeatureUnavailable
	}
	tenant := tenantKey(ctx, orgID)
	if tenant == "" {
		return nil, recmodel.ErrFeatureUnavailable
	}
	ns = normalizeNamespace(ns)

	cooc, ok, err := s.loadCooc(ctx, tenant, ns)
	if err != nil {
		return nil, err
	}
	if !ok && ns != "default" {
		ns = "default"
		cooc, ok, err = s.loadCooc(ctx, tenant, ns)
		if err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, recmodel.ErrFeatureUnavailable
	}
	out, err := s.selectCoocNeighbors(cooc, anchor, k)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 && ns != "default" {
		fallbackNS := "default"
		fallbackCooc, ok, err := s.loadCooc(ctx, tenant, fallbackNS)
		if err != nil {
			return nil, err
		}
		if ok {
			fallbackOut, err := s.selectCoocNeighbors(fallbackCooc, anchor, k)
			if err != nil {
				return nil, err
			}
			if len(fallbackOut) > 0 {
				return fallbackOut, nil
			}
		}
	}
	return out, nil
}

// ListItemsAvailability marks all requested items as available.
func (s *ArtifactAlgoStore) ListItemsAvailability(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		if id != "" {
			out[id] = true
		}
	}
	return out, nil
}

// CollaborativeTopK returns top implicit-feedback items for a user from artifacts.
func (s *ArtifactAlgoStore) CollaborativeTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	k int,
	excludeIDs []string,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		return nil, nil
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}
	if s == nil || s.loader == nil {
		return nil, recmodel.ErrFeatureUnavailable
	}
	tenant := tenantKey(ctx, orgID)
	if tenant == "" {
		return nil, recmodel.ErrFeatureUnavailable
	}
	ns = normalizeNamespace(ns)

	implicit, ok, err := s.loadImplicit(ctx, tenant, ns)
	if err != nil {
		return nil, err
	}
	if !ok && ns != "default" {
		ns = "default"
		implicit, ok, err = s.loadImplicit(ctx, tenant, ns)
		if err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, recmodel.ErrFeatureUnavailable
	}

	excludeSet := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			excludeSet[id] = struct{}{}
		}
	}
	var items []recmodel.ScoredItem
	for _, user := range implicit.Users {
		if strings.TrimSpace(user.UserID) != userID {
			continue
		}
		for _, item := range user.Items {
			if item.ItemID == "" || item.Score <= 0 {
				continue
			}
			if _, excluded := excludeSet[item.ItemID]; excluded {
				continue
			}
			items = append(items, recmodel.ScoredItem{ItemID: item.ItemID, Score: item.Score})
		}
		break
	}
	if len(items) == 0 {
		return nil, nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Score > items[j].Score
	})
	if k > 0 && len(items) > k {
		items = items[:k]
	}
	return items, nil
}

// ContentSimilarityTopK returns candidates by tag overlap using content artifacts.
func (s *ArtifactAlgoStore) ContentSimilarityTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	tags []string,
	k int,
	excludeIDs []string,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		return nil, nil
	}
	tags = normalizeTags(tags)
	if len(tags) == 0 {
		return nil, nil
	}
	if s == nil || s.loader == nil {
		return nil, recmodel.ErrFeatureUnavailable
	}
	tenant := tenantKey(ctx, orgID)
	if tenant == "" {
		return nil, recmodel.ErrFeatureUnavailable
	}
	ns = normalizeNamespace(ns)

	content, ok, err := s.loadContent(ctx, tenant, ns)
	if err != nil {
		return nil, err
	}
	if !ok && ns != "default" {
		ns = "default"
		content, ok, err = s.loadContent(ctx, tenant, ns)
		if err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, recmodel.ErrFeatureUnavailable
	}

	excludeSet := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			excludeSet[id] = struct{}{}
		}
	}
	tagSet := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if tag != "" {
			tagSet[tag] = struct{}{}
		}
	}

	items := make([]recmodel.ScoredItem, 0, len(content.Items))
	for _, item := range content.Items {
		id := strings.TrimSpace(item.ItemID)
		if id == "" {
			continue
		}
		if _, excluded := excludeSet[id]; excluded {
			continue
		}
		score := 0.0
		for _, tag := range normalizeTags(item.Tags) {
			if _, ok := tagSet[tag]; ok {
				score++
			}
		}
		if score <= 0 {
			continue
		}
		items = append(items, recmodel.ScoredItem{ItemID: id, Score: score})
	}
	if len(items) == 0 {
		return nil, nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Score > items[j].Score
	})
	if k > 0 && len(items) > k {
		items = items[:k]
	}
	return items, nil
}

// SessionSequenceTopK returns sequential candidates using session sequence artifacts.
func (s *ArtifactAlgoStore) SessionSequenceTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	lookback int,
	horizonMinutes float64,
	excludeIDs []string,
	k int,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		return nil, nil
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}
	if s == nil || s.loader == nil {
		return nil, recmodel.ErrFeatureUnavailable
	}
	_ = lookback
	_ = horizonMinutes

	tenant := tenantKey(ctx, orgID)
	if tenant == "" {
		return nil, recmodel.ErrFeatureUnavailable
	}
	ns = normalizeNamespace(ns)

	sess, ok, err := s.loadSessionSeq(ctx, tenant, ns)
	if err != nil {
		return nil, err
	}
	if !ok && ns != "default" {
		ns = "default"
		sess, ok, err = s.loadSessionSeq(ctx, tenant, ns)
		if err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, recmodel.ErrFeatureUnavailable
	}

	excludeSet := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			excludeSet[id] = struct{}{}
		}
	}
	var items []recmodel.ScoredItem
	for _, user := range sess.Users {
		if strings.TrimSpace(user.UserID) != userID {
			continue
		}
		for _, item := range user.Items {
			if item.ItemID == "" || item.Score <= 0 {
				continue
			}
			if _, excluded := excludeSet[item.ItemID]; excluded {
				continue
			}
			items = append(items, recmodel.ScoredItem{ItemID: item.ItemID, Score: item.Score})
		}
		break
	}
	if len(items) == 0 {
		return nil, nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Score > items[j].Score
	})
	if k > 0 && len(items) > k {
		items = items[:k]
	}
	return items, nil
}

func (s *ArtifactAlgoStore) loadPopularity(ctx context.Context, tenant, ns string) ([]artifacts.PopularityItem, bool, error) {
	if s == nil || s.loader == nil {
		return nil, false, nil
	}
	manifest, ok, err := s.loader.LoadManifest(ctx, tenant, ns)
	if err != nil || !ok {
		return nil, ok, err
	}
	uri := strings.TrimSpace(manifest.Current[artifacts.TypePopularity])
	if uri == "" {
		return nil, true, nil
	}
	pop, ok, err := s.loader.LoadPopularity(ctx, uri)
	if err != nil || !ok {
		return nil, ok, err
	}
	return pop.Items, true, nil
}

func (s *ArtifactAlgoStore) loadCooc(ctx context.Context, tenant, ns string) (artifacts.CoocArtifactV1, bool, error) {
	if s == nil || s.loader == nil {
		return artifacts.CoocArtifactV1{}, false, nil
	}
	manifest, ok, err := s.loader.LoadManifest(ctx, tenant, ns)
	if err != nil || !ok {
		return artifacts.CoocArtifactV1{}, ok, err
	}
	uri := strings.TrimSpace(manifest.Current[artifacts.TypeCooc])
	if uri == "" {
		return artifacts.CoocArtifactV1{}, false, nil
	}
	cooc, ok, err := s.loader.LoadCooc(ctx, uri)
	if err != nil || !ok {
		return artifacts.CoocArtifactV1{}, ok, err
	}
	return cooc, true, nil
}

func (s *ArtifactAlgoStore) loadImplicit(ctx context.Context, tenant, ns string) (artifacts.ImplicitArtifactV1, bool, error) {
	if s == nil || s.loader == nil {
		return artifacts.ImplicitArtifactV1{}, false, nil
	}
	manifest, ok, err := s.loader.LoadManifest(ctx, tenant, ns)
	if err != nil || !ok {
		return artifacts.ImplicitArtifactV1{}, ok, err
	}
	uri := strings.TrimSpace(manifest.Current[artifacts.TypeImplicit])
	if uri == "" {
		return artifacts.ImplicitArtifactV1{}, false, nil
	}
	implicit, ok, err := s.loader.LoadImplicit(ctx, uri)
	if err != nil || !ok {
		return artifacts.ImplicitArtifactV1{}, ok, err
	}
	return implicit, true, nil
}

func (s *ArtifactAlgoStore) loadContent(ctx context.Context, tenant, ns string) (artifacts.ContentArtifactV1, bool, error) {
	if s == nil || s.loader == nil {
		return artifacts.ContentArtifactV1{}, false, nil
	}
	manifest, ok, err := s.loader.LoadManifest(ctx, tenant, ns)
	if err != nil || !ok {
		return artifacts.ContentArtifactV1{}, ok, err
	}
	uri := strings.TrimSpace(manifest.Current[artifacts.TypeContentSim])
	if uri == "" {
		return artifacts.ContentArtifactV1{}, false, nil
	}
	content, ok, err := s.loader.LoadContent(ctx, uri)
	if err != nil || !ok {
		return artifacts.ContentArtifactV1{}, ok, err
	}
	return content, true, nil
}

func (s *ArtifactAlgoStore) loadSessionSeq(ctx context.Context, tenant, ns string) (artifacts.SessionSeqArtifactV1, bool, error) {
	if s == nil || s.loader == nil {
		return artifacts.SessionSeqArtifactV1{}, false, nil
	}
	manifest, ok, err := s.loader.LoadManifest(ctx, tenant, ns)
	if err != nil || !ok {
		return artifacts.SessionSeqArtifactV1{}, ok, err
	}
	uri := strings.TrimSpace(manifest.Current[artifacts.TypeSessionSeq])
	if uri == "" {
		return artifacts.SessionSeqArtifactV1{}, false, nil
	}
	sess, ok, err := s.loader.LoadSessionSeq(ctx, uri)
	if err != nil || !ok {
		return artifacts.SessionSeqArtifactV1{}, ok, err
	}
	return sess, true, nil
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func (s *ArtifactAlgoStore) filterPopularity(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	k int,
	items []artifacts.PopularityItem,
	c *recmodel.PopConstraints,
) ([]recmodel.ScoredItem, error) {
	excludeSet := make(map[string]struct{})
	var includeTags []string
	var minPrice, maxPrice *float64
	var createdAfter *time.Time
	if c != nil {
		for _, id := range c.ExcludeItemIDs {
			if id != "" {
				excludeSet[id] = struct{}{}
			}
		}
		includeTags = recmodel.NormalizeTags(c.IncludeTagsAny)
		minPrice = c.MinPrice
		maxPrice = c.MaxPrice
		createdAfter = c.CreatedAfter
	}
	needsTags := len(includeTags) > 0 || minPrice != nil || maxPrice != nil || createdAfter != nil
	var tags map[string]recmodel.ItemTags
	if needsTags {
		if s == nil || s.tagStore == nil {
			return nil, nil
		}
		ids := make([]string, 0, len(items))
		seen := make(map[string]struct{}, len(items))
		for _, item := range items {
			id := strings.TrimSpace(item.ItemID)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return nil, nil
		}
		var err error
		tags, err = s.tagStore.ListItemsTags(ctx, orgID, ns, ids)
		if err != nil {
			return nil, err
		}
	}

	out := make([]recmodel.ScoredItem, 0, k)
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ItemID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if _, excluded := excludeSet[id]; excluded {
			continue
		}
		if needsTags {
			info, ok := tags[id]
			if !ok {
				continue
			}
			if len(includeTags) > 0 {
				itemTags := recmodel.NormalizeTags(info.Tags)
				if !tagsIntersect(itemTags, includeTags) {
					continue
				}
			}
			if minPrice != nil {
				if info.Price == nil || *info.Price < *minPrice {
					continue
				}
			}
			if maxPrice != nil {
				if info.Price == nil || *info.Price > *maxPrice {
					continue
				}
			}
			if createdAfter != nil {
				if info.CreatedAt.IsZero() || info.CreatedAt.Before(*createdAfter) {
					continue
				}
			}
		}
		out = append(out, recmodel.ScoredItem{ItemID: id, Score: float64(item.Count)})
		if len(out) >= k {
			break
		}
	}
	return out, nil
}

func tagsIntersect(tags []string, required []string) bool {
	if len(tags) == 0 || len(required) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		set[tag] = struct{}{}
	}
	for _, req := range required {
		if req == "" {
			continue
		}
		if _, ok := set[req]; ok {
			return true
		}
	}
	return false
}

func (s *ArtifactAlgoStore) selectCoocNeighbors(cooc artifacts.CoocArtifactV1, anchor string, k int) ([]recmodel.ScoredItem, error) {
	if len(cooc.Neighbors) == 0 {
		return nil, nil
	}
	for _, row := range cooc.Neighbors {
		if row.ItemID != anchor {
			continue
		}
		if len(row.Items) == 0 {
			return nil, nil
		}
		sort.SliceStable(row.Items, func(i, j int) bool {
			if row.Items[i].Count == row.Items[j].Count {
				return row.Items[i].ItemID < row.Items[j].ItemID
			}
			return row.Items[i].Count > row.Items[j].Count
		})
		out := make([]recmodel.ScoredItem, 0, k)
		seen := make(map[string]struct{}, len(row.Items))
		for _, neighbor := range row.Items {
			id := strings.TrimSpace(neighbor.ItemID)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, recmodel.ScoredItem{ItemID: id, Score: float64(neighbor.Count)})
			if len(out) >= k {
				break
			}
		}
		return out, nil
	}
	return nil, nil
}

func normalizeNamespace(ns string) string {
	ns = strings.TrimSpace(ns)
	if ns == "" {
		return "default"
	}
	return ns
}

func tenantKey(ctx context.Context, orgID uuid.UUID) string {
	if ctx != nil {
		if v, ok := authorization.TenantIDFromContext(ctx); ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}
	if orgID != uuid.Nil {
		return orgID.String()
	}
	return ""
}

var _ recmodel.EngineStore = (*ArtifactAlgoStore)(nil)
var _ recmodel.CooccurrenceStore = (*ArtifactAlgoStore)(nil)
var _ recmodel.AvailabilityStore = (*ArtifactAlgoStore)(nil)
var _ recmodel.CollaborativeStore = (*ArtifactAlgoStore)(nil)
var _ recmodel.ContentStore = (*ArtifactAlgoStore)(nil)
var _ recmodel.SessionStore = (*ArtifactAlgoStore)(nil)
