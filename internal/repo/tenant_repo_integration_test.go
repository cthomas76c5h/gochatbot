package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/pagination"
	"gochatbot/internal/repo"
	"gochatbot/internal/testdb"
)

func TestTenantRepo_CreateAndGet(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	r := repo.NewTenantRepo(db.Conn)
	ctx := context.Background()

	created, err := r.Create(ctx, "Acme", "acme-law")
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	got, err := r.GetBySlug(ctx, "acme-law")
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Acme", got.Name)
}

func TestTenantRepo_UniqueSlug(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	r := repo.NewTenantRepo(db.Conn)
	ctx := context.Background()

	_, err := r.Create(ctx, "Acme", "acme-law")
	require.NoError(t, err)

	_, err = r.Create(ctx, "Acme2", "acme-law")
	require.Error(t, err)
}

func TestTenantRepo_ListWithCursor(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	r := repo.NewTenantRepo(db.Conn)
	ctx := context.Background()

	// Insert deterministic created_at ordering by sleeping a hair
	a, _ := r.Create(ctx, "A", "a")
	time.Sleep(10 * time.Millisecond)
	b, _ := r.Create(ctx, "B", "b")
	time.Sleep(10 * time.Millisecond)
	c, _ := r.Create(ctx, "C", "c")

	page1, cur, err := r.List(ctx, 2, nil)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.NotNil(t, cur)

	// should be newest first (c then b)
	require.Equal(t, c.ID, page1[0].ID)
	require.Equal(t, b.ID, page1[1].ID)

	page2, cur2, err := r.List(ctx, 2, &pagination.Cursor{CreatedAt: cur.CreatedAt, ID: cur.ID})
	require.NoError(t, err)
	require.Len(t, page2, 1)
	require.Nil(t, cur2)
	require.Equal(t, a.ID, page2[0].ID)
}
