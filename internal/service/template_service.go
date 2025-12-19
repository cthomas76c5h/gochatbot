package service

import (
	"context"
	"encoding/json"
	"errors"

	"gochatbot/internal/domain"
	"gochatbot/internal/httpapi"
	"gochatbot/internal/pagination"
	"gochatbot/internal/repo"
	"gochatbot/internal/validate"
)

type TemplateRepo interface {
	CreateTemplate(ctx context.Context, tenantID, name, slug string) (repo.Template, error)
	GetTemplateBySlug(ctx context.Context, tenantID, slug string) (repo.Template, error)
	ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) ([]repo.Template, *pagination.Cursor, error)

	CreateDraftVersion(ctx context.Context, templateID string, contentJSON []byte) (repo.TemplateVersion, error)
	PublishVersion(ctx context.Context, templateID string, version int) (repo.TemplateVersion, error)
	GetPublishedVersion(ctx context.Context, templateID string) (repo.TemplateVersion, error)
}

type TemplateService struct {
	repo TemplateRepo
}

func NewTemplateService(r TemplateRepo) *TemplateService {
	return &TemplateService{repo: r}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, tenantID, name, slug string) (httpapi.Template, error) {
	name = trim(name)
	if name == "" {
		return httpapi.Template{}, errors.New("invalid name")
	}

	norm, err := validate.NormalizeSlug(slug)
	if err != nil {
		return httpapi.Template{}, domain.ErrInvalidSlug
	}

	t, err := s.repo.CreateTemplate(ctx, tenantID, name, norm)
	if err != nil {
		return httpapi.Template{}, err
	}
	return httpapi.Template{ID: t.ID, TenantID: t.TenantID, Name: t.Name, Slug: t.Slug, CreatedAt: t.CreatedAt}, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, tenantID, slug string) (httpapi.Template, error) {
	norm, err := validate.NormalizeSlug(slug)
	if err != nil {
		return httpapi.Template{}, domain.ErrInvalidSlug
	}
	t, err := s.repo.GetTemplateBySlug(ctx, tenantID, norm)
	if err != nil {
		return httpapi.Template{}, err
	}
	return httpapi.Template{ID: t.ID, TenantID: t.TenantID, Name: t.Name, Slug: t.Slug, CreatedAt: t.CreatedAt}, nil
}

func (s *TemplateService) ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) (httpapi.ListTemplatesResult, error) {
	items, next, err := s.repo.ListTemplates(ctx, tenantID, limit, cursor)
	if err != nil {
		return httpapi.ListTemplatesResult{}, err
	}

	out := make([]httpapi.Template, 0, len(items))
	for _, t := range items {
		out = append(out, httpapi.Template{ID: t.ID, TenantID: t.TenantID, Name: t.Name, Slug: t.Slug, CreatedAt: t.CreatedAt})
	}
	var nextEnc string
	if next != nil {
		nextEnc = pagination.Encode(*next)
	}
	return httpapi.ListTemplatesResult{Items: out, NextCursor: nextEnc}, nil
}

func (s *TemplateService) CreateDraft(ctx context.Context, templateID string, content json.RawMessage) (httpapi.TemplateVersion, error) {
	v, err := s.repo.CreateDraftVersion(ctx, templateID, []byte(content))
	if err != nil {
		return httpapi.TemplateVersion{}, err
	}
	return httpapi.TemplateVersion{ID: v.ID, TemplateID: v.TemplateID, Version: v.Version, Status: v.Status, Content: json.RawMessage(v.Content), CreatedAt: v.CreatedAt}, nil
}

func (s *TemplateService) Publish(ctx context.Context, templateID string, version int) (httpapi.TemplateVersion, error) {
	if version <= 0 {
		return httpapi.TemplateVersion{}, errors.New("invalid version")
	}
	v, err := s.repo.PublishVersion(ctx, templateID, version)
	if err != nil {
		return httpapi.TemplateVersion{}, err
	}
	return httpapi.TemplateVersion{ID: v.ID, TemplateID: v.TemplateID, Version: v.Version, Status: v.Status, Content: json.RawMessage(v.Content), CreatedAt: v.CreatedAt}, nil
}

func (s *TemplateService) GetPublished(ctx context.Context, templateID string) (httpapi.TemplateVersion, error) {
	v, err := s.repo.GetPublishedVersion(ctx, templateID)
	if err != nil {
		return httpapi.TemplateVersion{}, err
	}
	return httpapi.TemplateVersion{ID: v.ID, TemplateID: v.TemplateID, Version: v.Version, Status: v.Status, Content: json.RawMessage(v.Content), CreatedAt: v.CreatedAt}, nil
}
