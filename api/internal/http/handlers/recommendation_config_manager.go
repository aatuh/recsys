package handlers

import (
	"sync/atomic"
	"time"
)

// RecommendationConfigMetadata captures provenance for config updates.
type RecommendationConfigMetadata struct {
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	Source    string    `json:"source,omitempty"`
	Notes     string    `json:"notes,omitempty"`
}

// RecommendationConfigSnapshot stores the config plus metadata.
type RecommendationConfigSnapshot struct {
	Config   RecommendationConfig         `json:"config"`
	Metadata RecommendationConfigMetadata `json:"metadata"`
}

// RecommendationConfigManager manages the live recommendation config.
type RecommendationConfigManager struct {
	value atomic.Value
}

// NewRecommendationConfigManager seeds the manager with the initial config.
func NewRecommendationConfigManager(initial RecommendationConfig, meta RecommendationConfigMetadata) *RecommendationConfigManager {
	if meta.UpdatedAt.IsZero() {
		meta.UpdatedAt = time.Now().UTC()
	}
	if meta.Source == "" {
		meta.Source = "env"
	}
	mgr := &RecommendationConfigManager{}
	mgr.value.Store(RecommendationConfigSnapshot{
		Config:   initial.Clone(),
		Metadata: meta,
	})
	return mgr
}

// Snapshot returns a copy of the current snapshot.
func (m *RecommendationConfigManager) Snapshot() RecommendationConfigSnapshot {
	snap, _ := m.value.Load().(RecommendationConfigSnapshot)
	return RecommendationConfigSnapshot{
		Config:   snap.Config.Clone(),
		Metadata: snap.Metadata,
	}
}

// Current returns a copy of the current config.
func (m *RecommendationConfigManager) Current() RecommendationConfig {
	snap, _ := m.value.Load().(RecommendationConfigSnapshot)
	return snap.Config.Clone()
}

// Update replaces the current config and returns the new snapshot.
func (m *RecommendationConfigManager) Update(cfg RecommendationConfig, meta RecommendationConfigMetadata) RecommendationConfigSnapshot {
	if meta.UpdatedAt.IsZero() {
		meta.UpdatedAt = time.Now().UTC()
	}
	if meta.Source == "" {
		meta.Source = "api"
	}
	snap := RecommendationConfigSnapshot{
		Config:   cfg.Clone(),
		Metadata: meta,
	}
	m.value.Store(snap)
	return snap
}
