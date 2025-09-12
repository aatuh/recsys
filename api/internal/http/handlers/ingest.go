package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"recsys/internal/bandit"
	"recsys/internal/http/common"
	"recsys/internal/http/store"
	"recsys/internal/http/types"

	"github.com/google/uuid"
)

// Keep in sync with store.EmbeddingDims.
const embeddingDims = 384

type Handler struct {
	Store                *store.Store
	DefaultOrg           uuid.UUID
	HalfLifeDays         float64
	PopularityWindowDays float64
	CoVisWindowDays      float64
	PopularityFanout     int
	MMRLambda            float64
	BrandCap             int
	CategoryCap          int
	RuleExcludePurchased bool
	PurchasedWindowDays  float64
	ProfileWindowDays    float64 // lookback for building profile; <=0 disables windowing
	ProfileBoost         float64 // multiplier in [0, +inf). 0 disables personalization
	ProfileTopNTags      int     // limit of profile tags considered
	BlendAlpha           float64
	BlendBeta            float64
	BlendGamma           float64
	BanditAlgo           bandit.Algorithm
}

// ItemsUpsert godoc
// @Summary      Upsert items (batch)
// @Description  Create or update items by opaque IDs.
// @Tags         ingestion
// @Accept       json
// @Produce      json
// @Param        payload  body  types.ItemsUpsertRequest  true  "Items upsert"
// @Success      202      {object}  types.Ack
// @Failure      400      {object}  common.APIError
// @Router       /v1/items:upsert [post]
// @ID upsertItems
func (h *Handler) ItemsUpsert(w http.ResponseWriter, r *http.Request) {
	var req types.ItemsUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}

	batch := make([]store.ItemUpsert, 0, len(req.Items))
	for _, it := range req.Items {
		// Validate embedding length if provided.
		var emb *[]float64
		if len(it.Embedding) > 0 {
			if len(it.Embedding) != embeddingDims {
				common.BadRequest(
					w, r,
					"embedding_dim_mismatch",
					"embedding length must be 384",
					map[string]any{"got": len(it.Embedding)},
				)
				return
			}
			tmp := make([]float64, len(it.Embedding))
			copy(tmp, it.Embedding)
			emb = &tmp
		}
		batch = append(batch, store.ItemUpsert{
			ItemID:    it.ItemID,
			Available: it.Available,
			Price:     it.Price,
			Tags:      it.Tags,
			Props:     it.Props,
			Embedding: emb,
		})
	}

	if err := h.Store.UpsertItems(r.Context(), orgID, req.Namespace, batch); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

// UsersUpsert godoc
// @Summary      Upsert users (batch)
// @Tags         ingestion
// @Accept       json
// @Produce      json
// @Param        payload  body  types.UsersUpsertRequest  true  "Users upsert"
// @Success      202      {object}  types.Ack
// @Failure      400      {object}  common.APIError
// @Router       /v1/users:upsert [post]
// @ID upsertUsers
func (h *Handler) UsersUpsert(w http.ResponseWriter, r *http.Request) {
	var req types.UsersUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}
	batch := make([]store.UserUpsert, 0, len(req.Users))
	for _, u := range req.Users {
		batch = append(batch, store.UserUpsert{UserID: u.UserID, Traits: u.Traits})
	}
	if err := h.Store.UpsertUsers(r.Context(), orgID, req.Namespace, batch); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

// EventsBatch godoc
// @Summary      Ingest events (batch)
// @Tags         ingestion
// @Accept       json
// @Produce      json
// @Param        payload  body  types.EventsBatchRequest  true  "Events batch"
// @Success      202      {object}  types.Ack
// @Failure      400      {object}  common.APIError
// @Router       /v1/events:batch [post]
// @ID batchEvents
func (h *Handler) EventsBatch(w http.ResponseWriter, r *http.Request) {
	var req types.EventsBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	orgID := h.DefaultOrg
	if s := r.Header.Get("X-Org-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			orgID = id
		}
	}
	batch := make([]store.EventInsert, 0, len(req.Events))
	for _, e := range req.Events {
		t := time.Now().UTC()
		if e.TS != "" {
			pt, err := time.Parse(time.RFC3339, e.TS)
			if err != nil {
				common.HttpError(w, r, err, http.StatusBadRequest)
				return
			}
			t = pt
		}
		batch = append(batch, store.EventInsert{
			UserID: e.UserID, ItemID: e.ItemID, Type: e.Type, Value: e.Value, TS: t, Meta: e.Meta, SourceEventID: e.SourceEventID})
	}
	if err := h.Store.InsertEvents(r.Context(), orgID, req.Namespace, batch); err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}
