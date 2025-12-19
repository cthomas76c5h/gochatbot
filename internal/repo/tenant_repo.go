package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
)

type Tenant struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
}

type TenantRepo struct {
	conn *pgx.Conn
}

func NewTenantRepo(conn *pgx.Conn) *TenantRepo {
	return &TenantRepo{conn: conn}
}

func (r *TenantRepo) Create(ctx context.Context, name, slug string) (Tenant, error) {
	var t Tenant
	err := r.conn.QueryRow(ctx, `
		insert into tenants (name, slug)
		values ($1, $2)
		returning id::text, name, slug, created_at
	`, name, slug).Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return Tenant{}, domain.ErrTenantSlugTaken
		}
		return Tenant{}, err
	}
	return t, nil
}

func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (Tenant, error) {
	var t Tenant
	err := r.conn.QueryRow(ctx, `
		select id::text, name, slug, created_at
		from tenants
		where slug = $1
	`, slug).Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenant{}, domain.ErrTenantNotFound
		}
		return Tenant{}, err
	}
	return t, nil
}

func (r *TenantRepo) List(ctx context.Context, limit int, cursor *pagination.Cursor) ([]Tenant, *pagination.Cursor, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Stable ordering: created_at DESC, id DESC
	var rows pgx.Rows
	var err error

	if cursor == nil {
		rows, err = r.conn.Query(ctx, `
			select id::text, name, slug, created_at
			from tenants
			order by created_at desc, id desc
			limit $1
		`, limit)
	} else {
		rows, err = r.conn.Query(ctx, `
			select id::text, name, slug, created_at
			from tenants
			where (created_at, id) < ($1::timestamptz, $2::uuid)
			order by created_at desc, id desc
			limit $3
		`, cursor.CreatedAt, cursor.ID, limit)
	}
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := make([]Tenant, 0, limit)
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, nil, err
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// next cursor = last row (if we returned a full page)
	if len(items) == limit {
		last := items[len(items)-1]
		next := &pagination.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
		return items, next, nil
	}

	return items, nil, nil
}


