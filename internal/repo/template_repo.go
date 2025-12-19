package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

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
	Status     string
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

	var (
		rows pgx.Rows
		err  error
	)
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
	if len(contentJSON) == 0 {
		contentJSON = []byte(`{}`)
	}
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

	cmdTag, err := r.conn.Exec(ctx, `
        update template_versions
        set status = 'published'
        where template_id = $1::uuid
          and version = $2
          and status = 'draft'
    `, templateID, version)
	if err != nil {
		if isUniqueViolation(err) {
			return TemplateVersion{}, domain.ErrVersionAlreadyPublished
		}
		return TemplateVersion{}, err
	}
	if cmdTag.RowsAffected() == 0 {
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


