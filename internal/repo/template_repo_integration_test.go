package repo_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
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

	v2, err := r.CreateDraftVersion(ctx, tpl.ID, []byte(`{"x":2}`))
	require.NoError(t, err)
	require.Equal(t, 2, v2.Version)

	_, err = r.PublishVersion(ctx, tpl.ID, 2)
	require.ErrorIs(t, err, domain.ErrVersionAlreadyPublished)
}
