// File: internal/http/handlers/foo.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"recsys/internal/http/mapper"
	"recsys/internal/services/foosvc"
	"recsys/internal/validation"
	endpointspec "recsys/src/specs/endpoints"
	"recsys/src/specs/types"

	"github.com/aatuh/api-toolkit-contrib/adapters/chi"
	listquery "github.com/aatuh/api-toolkit/endpoints/list"
	"github.com/aatuh/api-toolkit/httpx"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/aatuh/api-toolkit/response_writer"
)

// FooHandler exposes HTTP endpoints for Foo service.
type FooHandler struct {
	Svc       *foosvc.Service
	Logger    ports.Logger
	Validator ports.Validator
}

func NewFooHandler(
	s *foosvc.Service, l ports.Logger, v ports.Validator,
) *FooHandler {
	return &FooHandler{Svc: s, Logger: l, Validator: v}
}

// Routes returns a ports.HTTPRouter mounted by main.
func (h *FooHandler) Routes() ports.HTTPRouter {
	r := chi.New()
	r.Post(endpointspec.FooCreate, h.create)
	r.Get(endpointspec.FooList, h.list)
	r.Get(endpointspec.FooByID, h.get)
	r.Put(endpointspec.FooUpdate, h.update)
	r.Delete(endpointspec.FooDelete, h.delete)
	return r
}

// create handles POST /api/v1/foo
// @Summary Create foo
// @Description Create a new foo resource
// @Tags Foo
// @Accept json
// @Produce json
// @Param payload body types.CreateFooDTO true "Foo payload"
// @Success 201 {object} types.FooDTO
// @Failure 400 {object} types.Problem
// @Failure 409 {object} types.Problem
// @Router /api/v1/foo [post]
func (h *FooHandler) create(w http.ResponseWriter, r *http.Request) {
	var dto types.CreateFooDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}
	// First, struct-tag validator hooks (no-op if basic)
	if err := h.Validator.ValidateStruct(r.Context(), &dto); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	// Then, custom rules
	if err := validation.ValidateCreateFoo(&dto); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	f, err := h.Svc.Create(r.Context(), mapper.CreateFooInput(&dto))
	if err != nil {
		writeDomainErr(w, err)
		return
	}

	response_writer.WriteJSON(w, http.StatusCreated, mapper.FooDTOFromModel(f))
}

// get handles GET /api/v1/foo/{id}
// @Summary Get foo
// @Description Fetch a foo by ID
// @Tags Foo
// @Produce json
// @Param id path string true "Foo ID"
// @Success 200 {object} types.FooDTO
// @Failure 404 {object} types.Problem
// @Router /api/v1/foo/{id} [get]
func (h *FooHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	f, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeDomainErr(w, err)
		return
	}
	response_writer.WriteJSON(w, http.StatusOK, mapper.FooDTOFromModel(f))
}

// update handles PUT /api/v1/foo/{id}
// @Summary Update foo
// @Description Update a foo by ID
// @Tags Foo
// @Accept json
// @Produce json
// @Param id path string true "Foo ID"
// @Param payload body types.UpdateFooDTO true "Foo update payload"
// @Success 200 {object} types.FooDTO
// @Failure 400 {object} types.Problem
// @Failure 404 {object} types.Problem
// @Router /api/v1/foo/{id} [put]
func (h *FooHandler) update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	var dto types.UpdateFooDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}
	if err := h.Validator.ValidateStruct(r.Context(), &dto); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidateUpdateFoo(&dto); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	f, err := h.Svc.Update(r.Context(), mapper.UpdateFooInput(&dto, id))
	if err != nil {
		writeDomainErr(w, err)
		return
	}

	response_writer.WriteJSON(w, http.StatusOK, mapper.FooDTOFromModel(f))
}

// delete handles DELETE /api/v1/foo/{id}
// @Summary Delete foo
// @Description Delete a foo by ID
// @Tags Foo
// @Param id path string true "Foo ID"
// @Success 204 "No Content"
// @Failure 404 {object} types.Problem
// @Router /api/v1/foo/{id} [delete]
func (h *FooHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Svc.Delete(r.Context(), id); err != nil {
		writeDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// list handles GET /api/v1/foo
// @Summary List foos
// @Description List foos filtered by org and namespace
// @Tags Foo
// @Produce json
// @Param org_id query string true "Organization ID"
// @Param namespace query string true "Namespace"
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Param search query string false "Search term"
// @Success 200 {object} types.FooListResponse
// @Failure 400 {object} types.Problem
// @Router /api/v1/foo [get]
func (h *FooHandler) list(w http.ResponseWriter, r *http.Request) {
	q := listquery.ParseListQuery(r, listquery.ListQueryConfig{
		DefaultLimit:   50,
		MaxLimit:       200,
		AllowedFilters: []string{"org_id", "namespace"},
		Required:       []string{"org_id", "namespace"},
	})
	if missing := q.MissingRequired(); len(missing) > 0 {
		writeErr(w, http.StatusBadRequest,
			"missing required filters: "+strings.Join(missing, ", "))
		return
	}

	orgID := q.First("org_id")
	ns := q.First("namespace")

	res, err := h.Svc.List(r.Context(), orgID, ns,
		q.Limit, q.Offset, q.Search)
	if err != nil {
		writeDomainErr(w, err)
		return
	}

	items := make([]types.FooDTO, len(res.Items))
	for i := range res.Items {
		items[i] = mapper.FooDTOFromModel(&res.Items[i])
	}

	meta := types.ListMeta{
		Total:  res.Total,
		Count:  len(items),
		Limit:  q.Limit,
		Offset: q.Offset,
		Search: q.Search,
	}
	if len(q.Filters) > 0 {
		meta.Filters = cloneFilterMap(q.Filters)
	}

	out := types.FooListResponse{
		Data: items,
		Meta: meta,
	}
	response_writer.WriteJSON(w, http.StatusOK, out)
}

func cloneFilterMap(in listquery.Filters) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for k, vals := range in {
		cp := make([]string, len(vals))
		copy(cp, vals)
		out[k] = cp
	}
	return out
}

func writeDomainErr(w http.ResponseWriter, err error) {
	switch err {
	case foosvc.ErrInvalid:
		httpx.WriteProblem(w, http.StatusBadRequest, httpx.Problem{
			Title:  http.StatusText(http.StatusBadRequest),
			Detail: err.Error(),
		})
	case foosvc.ErrNotFound:
		httpx.WriteProblem(w, http.StatusNotFound, httpx.Problem{
			Title:  http.StatusText(http.StatusNotFound),
			Detail: err.Error(),
		})
	case foosvc.ErrConflict:
		httpx.WriteProblem(w, http.StatusConflict, httpx.Problem{
			Title:  http.StatusText(http.StatusConflict),
			Detail: err.Error(),
		})
	default:
		httpx.WriteProblem(w, http.StatusInternalServerError, httpx.Problem{
			Title:  http.StatusText(http.StatusInternalServerError),
			Detail: "internal error",
		})
	}
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	httpx.WriteProblem(w, code, httpx.Problem{
		Title:  http.StatusText(code),
		Detail: msg,
	})
}
