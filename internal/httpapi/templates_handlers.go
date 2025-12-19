package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
	"gochatbot/internal/validate"
)

type Template struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type ListTemplatesResult struct {
	Items      []Template `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

type TemplateVersion struct {
	ID         string          `json:"id"`
	TemplateID string          `json:"template_id"`
	Version    int             `json:"version"`
	Status     string          `json:"status"`
	Content    json.RawMessage `json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
}

type TemplateService interface {
	CreateTemplate(ctx context.Context, tenantID, name, slug string) (Template, error)
	GetTemplate(ctx context.Context, tenantID, slug string) (Template, error)
	ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) (ListTemplatesResult, error)
	CreateDraft(ctx context.Context, templateID string, content json.RawMessage) (TemplateVersion, error)
	Publish(ctx context.Context, templateID string, version int) (TemplateVersion, error)
	GetPublished(ctx context.Context, templateID string) (TemplateVersion, error)
}

type createTemplateReq struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantSlug := chi.URLParam(r, "tenantSlug")
	tenantSlug, err := validate.NormalizeSlug(tenantSlug)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}

	tenant, err := s.deps.TenantSvc.GetTenantBySlug(RequestContext{}, tenantSlug)
	if err != nil {
		if errors.Is(err, domain.ErrTenantNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "tenant not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	var req createTemplateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	req.Name = trim(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "invalid name"})
		return
	}

	slug, err := validate.NormalizeSlug(req.Slug)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}

	tpl, err := s.deps.TemplateSvc.CreateTemplate(r.Context(), tenant.ID, req.Name, slug)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTemplateSlugTaken):
			writeJSON(w, http.StatusConflict, map[string]any{"error": "slug taken"})
			return
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
			return
		}
	}

	writeJSON(w, http.StatusCreated, tpl)
}

func (s *Server) handleGetTemplateBySlug(w http.ResponseWriter, r *http.Request) {
	tenantSlug := chi.URLParam(r, "tenantSlug")
	tenantSlug, err := validate.NormalizeSlug(tenantSlug)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}
	templateSlug := chi.URLParam(r, "templateSlug")
	templateSlug, err = validate.NormalizeSlug(templateSlug)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}

	tenant, err := s.deps.TenantSvc.GetTenantBySlug(RequestContext{}, tenantSlug)
	if err != nil {
		if errors.Is(err, domain.ErrTenantNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "tenant not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	tpl, err := s.deps.TemplateSvc.GetTemplate(r.Context(), tenant.ID, templateSlug)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	writeJSON(w, http.StatusOK, tpl)
}

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	tenantSlug := chi.URLParam(r, "tenantSlug")
	tenantSlug, err := validate.NormalizeSlug(tenantSlug)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": domain.ErrInvalidSlug.Error()})
		return
	}

	tenant, err := s.deps.TenantSvc.GetTenantBySlug(RequestContext{}, tenantSlug)
	if err != nil {
		if errors.Is(err, domain.ErrTenantNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "tenant not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

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

	res, err := s.deps.TemplateSvc.ListTemplates(r.Context(), tenant.ID, limit, cur)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

type createDraftReq struct {
	Content json.RawMessage `json:"content"`
}

func (s *Server) handleCreateDraft(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateID")
	if strings.TrimSpace(templateID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid template id"})
		return
	}

	var req createDraftReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	v, err := s.deps.TemplateSvc.CreateDraft(r.Context(), templateID, req.Content)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}

	writeJSON(w, http.StatusCreated, v)
}

type publishReq struct {
	Version int `json:"version"`
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateID")
	if strings.TrimSpace(templateID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid template id"})
		return
	}

	var req publishReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if req.Version <= 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "invalid version"})
		return
	}

	v, err := s.deps.TemplateSvc.Publish(r.Context(), templateID, req.Version)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVersionNotFound):
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		case errors.Is(err, domain.ErrVersionAlreadyPublished):
			writeJSON(w, http.StatusConflict, map[string]any{"error": "already published"})
			return
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
			return
		}
	}

	writeJSON(w, http.StatusOK, v)
}

func (s *Server) handleGetPublished(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateID")
	if strings.TrimSpace(templateID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid template id"})
		return
	}

	v, err := s.deps.TemplateSvc.GetPublished(r.Context(), templateID)
	if err != nil {
		if errors.Is(err, domain.ErrVersionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
		return
	}
	writeJSON(w, http.StatusOK, v)
}
