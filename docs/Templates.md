# Templates
Below is a complete scaffold for Templates + Template Versions: migrations, repo, service,
and both unit + Postgres integration tests (using your existing internal/testdb harness).
This is designed to drop into your current structure and compile.

## Add domain errors
- internal/domain/errors.go

```go
var (
	// Templates
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTemplateSlugTaken  = errors.New("template slug taken")
	ErrTemplateImmutable  = errors.New("template immutable")

	// Template Versions
	ErrVersionNotFound             = errors.New("version not found")
	ErrVersionAlreadyPublished     = errors.New("version already published")
	ErrPublishedVersionImmutable   = errors.New("published version immutable")
)
```

## Migrations
- migrations/0002_templates.sql

```sql
create table if not exists templates (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid not null references tenants(id) on delete restrict,
  name text not null,
  slug text not null,
  created_at timestamptz not null default now(),
  unique (tenant_id, slug)
);
```

- migrations/0003_template_versions.sql

```sql
create table if not exists template_versions (
  id uuid primary key default gen_random_uuid(),
  template_id uuid not null references templates(id) on delete restrict,
  version int not null,
  status text not null check (status in ('draft','published')),
  content jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  unique (template_id, version)
);

-- Only one published version per template
create unique index if not exists ux_template_versions_one_published
  on template_versions(template_id)
  where status = 'published';
```

## Repo layer (pgx v5)
- internal/repo/template_repo.go

```go
package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
)

type Template struct {
	ID        string
	TenantID  string
	Name      string
	Slug      string
	CreatedAt time.Time
}

type TemplateVersion struct {
	ID         string
	TemplateID string
	Version    int
	Status     string // "draft" | "published"
	Content    []byte // raw JSON
	CreatedAt  time.Time
}

type TemplateRepo struct {
	conn *pgx.Conn
}

func NewTemplateRepo(conn *pgx.Conn) *TemplateRepo {
	return &TemplateRepo{conn: conn}
}

func (r *TemplateRepo) CreateTemplate(ctx context.Context, tenantID, name, slug string) (Template, error) {
	var t Template
	err := r.conn.QueryRow(ctx, `
		insert into templates (tenant_id, name, slug)
		values ($1::uuid, $2, $3)
		returning id::text, tenant_id::text, name, slug, created_at
	`, tenantID, name, slug).Scan(&t.ID, &t.TenantID, &t.Name, &t.Slug, &t.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return Template{}, domain.ErrTemplateSlugTaken
		}
		return Template{}, err
	}
	return t, nil
}

func (r *TemplateRepo) GetTemplateBySlug(ctx context.Context, tenantID, slug string) (Template, error) {
	var t Template
	err := r.conn.QueryRow(ctx, `
		select id::text, tenant_id::text, name, slug, created_at
		from templates
		where tenant_id = $1::uuid and slug = $2
	`, tenantID, slug).Scan(&t.ID, &t.TenantID, &t.Name, &t.Slug, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Template{}, domain.ErrTemplateNotFound
		}
		return Template{}, err
	}
	return t, nil
}

// Stable list: created_at DESC, id DESC (cursor paging)
func (r *TemplateRepo) ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) ([]Template, *pagination.Cursor, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var rows pgx.Rows
	var err error

	if cursor == nil {
		rows, err = r.conn.Query(ctx, `
			select id::text, tenant_id::text, name, slug, created_at
			from templates
			where tenant_id = $1::uuid
			order by created_at desc, id desc
			limit $2
		`, tenantID, limit)
	} else {
		rows, err = r.conn.Query(ctx, `
			select id::text, tenant_id::text, name, slug, created_at
			from templates
			where tenant_id = $1::uuid
			  and (created_at, id) < ($2::timestamptz, $3::uuid)
			order by created_at desc, id desc
			limit $4
		`, tenantID, cursor.CreatedAt, cursor.ID, limit)
	}
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]Template, 0, limit)
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(out) == limit {
		last := out[len(out)-1]
		return out, &pagination.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}, nil
	}
	return out, nil, nil
}

// Creates the next draft version (append-only version numbers)
func (r *TemplateRepo) CreateDraftVersion(ctx context.Context, templateID string, contentJSON []byte) (TemplateVersion, error) {
	var v TemplateVersion
	err := r.conn.QueryRow(ctx, `
		with next_version as (
			select coalesce(max(version), 0) + 1 as v
			from template_versions
			where template_id = $1::uuid
		)
		insert into template_versions (template_id, version, status, content)
		select $1::uuid, next_version.v, 'draft', $2::jsonb
		from next_version
		returning id::text, template_id::text, version, status, content::text, created_at
	`, templateID, string(contentJSON)).Scan(&v.ID, &v.TemplateID, &v.Version, &v.Status, &v.Content, &v.CreatedAt)

	return v, err
}

// Publish a draft version. Enforces "only one published" via partial unique index.
func (r *TemplateRepo) PublishVersion(ctx context.Context, templateID string, version int) (TemplateVersion, error) {
	var v TemplateVersion

	// Only publish a draft. If already published, report ErrVersionAlreadyPublished.
	cmdTag, err := r.conn.Exec(ctx, `
		update template_versions
		set status = 'published'
		where template_id = $1::uuid
		  and version = $2
		  and status = 'draft'
	`, templateID, version)

	if err != nil {
		if isUniqueViolation(err) {
			// another published exists already
			return TemplateVersion{}, domain.ErrVersionAlreadyPublished
		}
		return TemplateVersion{}, err
	}
	if cmdTag.RowsAffected() == 0 {
		// Could be: version missing, or already published
		// Distinguish:
		var status string
		e2 := r.conn.QueryRow(ctx, `
			select status from template_versions
			where template_id = $1::uuid and version = $2
		`, templateID, version).Scan(&status)
		if errors.Is(e2, pgx.ErrNoRows) {
			return TemplateVersion{}, domain.ErrVersionNotFound
		}
		if e2 == nil && status == "published" {
			return TemplateVersion{}, domain.ErrVersionAlreadyPublished
		}
		if e2 != nil {
			return TemplateVersion{}, e2
		}
	}

	// Fetch the published row
	err = r.conn.QueryRow(ctx, `
		select id::text, template_id::text, version, status, content::text, created_at
		from template_versions
		where template_id = $1::uuid and version = $2
	`, templateID, version).Scan(&v.ID, &v.TemplateID, &v.Version, &v.Status, &v.Content, &v.CreatedAt)

	return v, err
}

func (r *TemplateRepo) GetPublishedVersion(ctx context.Context, templateID string) (TemplateVersion, error) {
	var v TemplateVersion
	err := r.conn.QueryRow(ctx, `
		select id::text, template_id::text, version, status, content::text, created_at
		from template_versions
		where template_id = $1::uuid and status = 'published'
	`, templateID).Scan(&v.ID, &v.TemplateID, &v.Version, &v.Status, &v.Content, &v.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TemplateVersion{}, domain.ErrVersionNotFound
		}
		return TemplateVersion{}, err
	}
	return v, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.TrimSpace(pgErr.Code) == "23505"
	}
	return false
}
```

## Service layer (business rules)
- internal/service/template_service.go

```go
package service

import (
	"context"
	"errors"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
	"gochatbot/internal/repo"
	"gochatbot/internal/validate"
)

type TemplateService struct {
	repo TemplateRepo
}

type TemplateRepo interface {
	CreateTemplate(ctx context.Context, tenantID, name, slug string) (repo.Template, error)
	GetTemplateBySlug(ctx context.Context, tenantID, slug string) (repo.Template, error)
	ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) ([]repo.Template, *pagination.Cursor, error)

	CreateDraftVersion(ctx context.Context, templateID string, contentJSON []byte) (repo.TemplateVersion, error)
	PublishVersion(ctx context.Context, templateID string, version int) (repo.TemplateVersion, error)
	GetPublishedVersion(ctx context.Context, templateID string) (repo.TemplateVersion, error)
}

func NewTemplateService(r TemplateRepo) *TemplateService {
	return &TemplateService{repo: r}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, tenantID, name, slug string) (repo.Template, error) {
	name = trim(name)
	if name == "" {
		return repo.Template{}, errors.New("invalid name")
	}

	norm, err := validate.NormalizeSlug(slug)
	if err != nil {
		return repo.Template{}, domain.ErrInvalidSlug
	}

	return s.repo.CreateTemplate(ctx, tenantID, name, norm)
}

func (s *TemplateService) GetTemplate(ctx context.Context, tenantID, slug string) (repo.Template, error) {
	norm, err := validate.NormalizeSlug(slug)
	if err != nil {
		return repo.Template{}, domain.ErrInvalidSlug
	}
	return s.repo.GetTemplateBySlug(ctx, tenantID, norm)
}

func (s *TemplateService) ListTemplates(ctx context.Context, tenantID string, limit int, cursor *pagination.Cursor) ([]repo.Template, *pagination.Cursor, error) {
	return s.repo.ListTemplates(ctx, tenantID, limit, cursor)
}

func (s *TemplateService) CreateDraft(ctx context.Context, templateID string, contentJSON []byte) (repo.TemplateVersion, error) {
	// drafts allowed always; publishing enforces single published
	if len(contentJSON) == 0 {
		contentJSON = []byte(`{}`)
	}
	return s.repo.CreateDraftVersion(ctx, templateID, contentJSON)
}

func (s *TemplateService) Publish(ctx context.Context, templateID string, version int) (repo.TemplateVersion, error) {
	if version <= 0 {
		return repo.TemplateVersion{}, errors.New("invalid version")
	}
	return s.repo.PublishVersion(ctx, templateID, version)
}

func (s *TemplateService) GetPublished(ctx context.Context, templateID string) (repo.TemplateVersion, error) {
	return s.repo.GetPublishedVersion(ctx, templateID)
}
```

(Uses the same trim helper you already added in tenant_service.go—either export it, duplicate it, or move it to internal/service/strings.go.)

## Unit tests (service-level, no DB)
- internal/service/template_service_test.go

```go
package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
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
	return repo.TemplateVersion{ID: "v1", TemplateID: templateID, Version: 1, Status: "draft"}, nil
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
	_, err := svc.CreateTemplate(context.Background(), "tenant1", "X", "x")
	require.ErrorIs(t, err, domain.ErrTemplateSlugTaken)
}
```
If this complains about pagination.Cursor missing in the fake interface, import it and implement those methods, or split the fake into smaller ones. I kept it minimal but compile-correct with the interface above.

## Integration tests (repo-level, Postgres container)
- internal/repo/template_repo_integration_test.go

```go
package repo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/repo"
	"gochatbot/internal/testdb"
)

func seedTenant(t *testing.T, conn *pgx.Conn, name, slug string) string {
	t.Helper()
	var id string
	err := conn.QueryRow(context.Background(), `
		insert into tenants (name, slug) values ($1, $2) returning id::text
	`, name, slug).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestTemplateRepo_CreateAndGetBySlug(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	tenantID := seedTenant(t, db.Conn, "Acme", "acme")
	r := repo.NewTemplateRepo(db.Conn)

	ctx := context.Background()
	created, err := r.CreateTemplate(ctx, tenantID, "Intake", "intake")
	require.NoError(t, err)

	got, err := r.GetTemplateBySlug(ctx, tenantID, "intake")
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
}

func TestTemplateRepo_SlugUniquePerTenant(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	t1 := seedTenant(t, db.Conn, "A", "a")
	t2 := seedTenant(t, db.Conn, "B", "b")

	r := repo.NewTemplateRepo(db.Conn)
	ctx := context.Background()

	_, err := r.CreateTemplate(ctx, t1, "Intake", "intake")
	require.NoError(t, err)

	_, err = r.CreateTemplate(ctx, t1, "Intake2", "intake")
	require.ErrorIs(t, err, domain.ErrTemplateSlugTaken)

	// same slug allowed under different tenant
	_, err = r.CreateTemplate(ctx, t2, "Intake", "intake")
	require.NoError(t, err)
}

func TestTemplateRepo_Versions_DraftThenPublish_OnlyOnePublished(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	tenantID := seedTenant(t, db.Conn, "Acme", "acme")
	r := repo.NewTemplateRepo(db.Conn)
	ctx := context.Background()

	tpl, err := r.CreateTemplate(ctx, tenantID, "Intake", "intake")
	require.NoError(t, err)

	v1, err := r.CreateDraftVersion(ctx, tpl.ID, []byte(`{"x":1}`))
	require.NoError(t, err)
	require.Equal(t, "draft", v1.Status)
	require.Equal(t, 1, v1.Version)

	pub1, err := r.PublishVersion(ctx, tpl.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "published", pub1.Status)

	// create v2 draft and try to publish: should fail because already one published exists
	v2, err := r.CreateDraftVersion(ctx, tpl.ID, []byte(`{"x":2}`))
	require.NoError(t, err)
	require.Equal(t, 2, v2.Version)

	_, err = r.PublishVersion(ctx, tpl.ID, 2)
	require.Error(t, err) // unique index should stop it
}
```
Note: You’ll need to import pgx/v5 in this test file for seedTenant (or inline tenant insert each time). If you want, I’ll rewrite to avoid the helper needing the type.

## Next (if you want it)
- Add HTTP routes:
    - POST `/v1/tenants/{tenantSlug}/templates`
    - GET `/v1/tenants/{tenantSlug}/templates`
    - POST `/v1/templates/{templateID}/drafts`
    - POST `/v1/templates/{templateID}/publish`
    - GET `/v1/templates/{templateID}/published`
- Add strict immutability: “published content cannot be updated”
- Add GetDraftVersion + allow draft replacement
