package handlers

import (
	"encoding/json"
	"net/http"

	"recsys/internal/http/common"
	"recsys/internal/http/types"

	"github.com/go-chi/chi/v5"
)

// Recommend godoc
// @Summary      Get recommendations for a user
// @Tags         ranking
// @Accept       json
// @Produce      json
// @Param        payload  body  types.RecommendRequest  true  "Recommend request"
// @Success      200      {object}  types.RecommendResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/recommendations [post]
func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req types.RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	// TODO: real scoring; stub a response to keep DX tight
	resp := types.RecommendResponse{
		ModelVersion: "pop_0",
		Items:        []types.ScoredItem{},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// ItemSimilar godoc
// @Summary      Get similar items
// @Tags         ranking
// @Produce      json
// @Param        item_id  path  string  true  "Item ID"
// @Param        k        query int     false "Top-K"  default(20)
// @Success      200      {array}  types.ScoredItem
// @Failure      400      {object} common.APIError
// @Router       /v1/items/{item_id}/similar [get]
func (h *Handler) ItemSimilar(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "item_id")
	// TODO: read k, fetch cooc results
	_ = json.NewEncoder(w).Encode([]types.ScoredItem{})
}
