package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
	"gochatbot/internal/repo"
	"gochatbot/internal/service"
)

type fakeTemplateRepo struct {
	createErr error
}

func (f *fakeTemplateRepo) CreateTemplate(ctx context.Context, tenantID, name, slug string) (repo.Template, error) {
	if f.createErr != nil {
		return repo.Template{}, f.createErr
	}
	return repo.Template{ID: "tpl1", TenantID: tenantID, Name: name, Slug: slug}, nil
}

func (f *fakeTemplateRepo) GetTemplateBySlug(ctx context.Context, tenantID, slug string) (repo.Template, error) {
	return repo.Template{}, domain.ErrTemplateNotFound
}

func (f *fakeTemplateRepo) ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) ([]repo.Template, *pagination.Cursor, error) {
	return nil, nil, nil
}

func (f *fakeTemplateRepo) CreateDraftVersion(ctx context.Context, templateID string, contentJSON []byte) (repo.TemplateVersion, error) {
	return repo.TemplateVersion{ID: "v1", TemplateID: templateID, Version: 1, Status: "draft", Content: contentJSON}, nil
}

func (f *fakeTemplateRepo) PublishVersion(ctx context.Context, templateID string, version int) (repo.TemplateVersion, error) {
	return repo.TemplateVersion{ID: "v1", TemplateID: templateID, Version: version, Status: "published"}, nil
}

func (f *fakeTemplateRepo) GetPublishedVersion(ctx context.Context, templateID string) (repo.TemplateVersion, error) {
	return repo.TemplateVersion{}, domain.ErrVersionNotFound
}

func TestTemplateService_CreateTemplate_NormalizesSlug(t *testing.T) {
	svc := service.NewTemplateService(&fakeTemplateRepo{})
	tpl, err := svc.CreateTemplate(context.Background(), "tenant1", "My Template", "My Template!!")
	require.NoError(t, err)
	require.Equal(t, "my-template", tpl.Slug)
}

func TestTemplateService_CreateTemplate_InvalidSlug(t *testing.T) {
	svc := service.NewTemplateService(&fakeTemplateRepo{})
	_, err := svc.CreateTemplate(context.Background(), "tenant1", "X", "!!")
	require.ErrorIs(t, err, domain.ErrInvalidSlug)
}

func TestTemplateService_CreateTemplate_ConflictPropagates(t *testing.T) {
	svc := service.NewTemplateService(&fakeTemplateRepo{createErr: domain.ErrTemplateSlugTaken})
	_, err := svc.CreateTemplate(context.Background(), "tenant1", "X", "xxx")
	require.ErrorIs(t, err, domain.ErrTemplateSlugTaken)
}

func TestTemplateService_CreateDraft_PassesJSONThrough(t *testing.T) {
	svc := service.NewTemplateService(&fakeTemplateRepo{})
	raw := json.RawMessage(`{"k":1}`)
	v, err := svc.CreateDraft(context.Background(), "tpl1", raw)
	require.NoError(t, err)
	require.JSONEq(t, string(raw), string(v.Content))
}
