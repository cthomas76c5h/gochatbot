package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
	"gochatbot/internal/validate"
)

type Tenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ListTenantsResult struct {
	Items      []Tenant `json:"items"`
	NextCursor string   `json:"next_cursor,omitempty"`
}

type TenantService interface {
	CreateTenant(rctx RequestContext, name string, slug string) (Tenant, error)
	GetTenantBySlug(rctx RequestContext, slug string) (Tenant, error)
	ListTenants(rctx RequestContext, limit int, cursor *pagination.Cursor) (ListTenantsResult, error)
}

type RequestContext struct {
	// later: Auth info, tenant scope, request id
}

type createTenantReq struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req createTenantReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	req.Name = trim(req.Name)
	slug, err := validate.NormalizeSlug(req.Slug)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "invalid name"})
		return
	}

	t, err := s.deps.TenantSvc.CreateTenant(RequestContext{}, req.Name, slug)
	if err != nil {
		if err == domain.ErrTenantSlugTaken {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "slug taken"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleGetTenantBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	slug, err := validate.NormalizeSlug(slug)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}

	t, err := s.deps.TenantSvc.GetTenantBySlug(RequestContext{}, slug)
	if err != nil {
		if errors.Is(err, domain.ErrTenantNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseLimit(r, 50, 1, 200)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
		return
	}

	var cur *pagination.Cursor
	if raw := strings.TrimSpace(r.URL.Query().Get("cursor")); raw != "" {
		decoded, err := pagination.Decode(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": domain.ErrInvalidCursor.Error()})
			return
		}
		cur = &decoded
	}

	res, err := s.deps.TenantSvc.ListTenants(RequestContext{}, limit, cur)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	writeJSON(w, http.StatusOK, res)
}
