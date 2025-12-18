package service

import (
	"context"

	"gochatbot/internal/domain"
	"gochatbot/internal/httpapi"
	"gochatbot/internal/pagination"
	"gochatbot/internal/repo"
	"gochatbot/internal/validate"
)

type TenantRepo interface {
	Create(ctx context.Context, name, slug string) (repo.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (repo.Tenant, error)
	List(ctx context.Context, limit int, cursor *pagination.Cursor) ([]repo.Tenant, *pagination.Cursor, error)
}

type TenantService struct {
	repo TenantRepo
}

func NewTenantService(r TenantRepo) *TenantService {
	return &TenantService{repo: r}
}

func (s *TenantService) CreateTenant(_ httpapi.RequestContext, name string, slug string) (httpapi.Tenant, error) {
	name = trim(name)
	if name == "" {
		return httpapi.Tenant{}, domain.ErrEmptyMessage
	}

	// Always normalize slug here too, even if handler did it (defense-in-depth).
	norm, err := validate.NormalizeSlug(slug)
	if err != nil {
		return httpapi.Tenant{}, domain.ErrInvalidSlug
	}

	t, err := s.repo.Create(context.Background(), name, norm)
	if err != nil {
		// pass through known domain errors
		if err == domain.ErrTenantSlugTaken {
			return httpapi.Tenant{}, err
		}
		return httpapi.Tenant{}, err
	}

	return httpapi.Tenant{ID: t.ID, Name: t.Name, Slug: t.Slug}, nil
}

func (s *TenantService) GetTenantBySlug(_ httpapi.RequestContext, slug string) (httpapi.Tenant, error) {
	norm, err := validate.NormalizeSlug(slug)
	if err != nil {
		return httpapi.Tenant{}, domain.ErrInvalidSlug
	}

	t, err := s.repo.GetBySlug(context.Background(), norm)
	if err != nil {
		return httpapi.Tenant{}, err
	}
	return httpapi.Tenant{ID: t.ID, Name: t.Name, Slug: t.Slug}, nil
}

func (s *TenantService) ListTenants(_ httpapi.RequestContext, limit int, cursor *pagination.Cursor) (httpapi.ListTenantsResult, error) {
	items, next, err := s.repo.List(context.Background(), limit, cursor)
	if err != nil {
		return httpapi.ListTenantsResult{}, err
	}

	out := make([]httpapi.Tenant, 0, len(items))
	for _, t := range items {
		out = append(out, httpapi.Tenant{ID: t.ID, Name: t.Name, Slug: t.Slug})
	}

	var nextEnc string
	if next != nil {
		nextEnc = pagination.Encode(*next)
	}

	return httpapi.ListTenantsResult{Items: out, NextCursor: nextEnc}, nil
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}
