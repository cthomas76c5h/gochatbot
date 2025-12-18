package repo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/testdb"
)

func TestTenants_UniqueSlug(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	ctx := context.Background()

	_, err := db.Conn.Exec(ctx, `
		insert into tenants (name, slug) values ($1, $2)
	`, "Acme", "acme-law")
	require.NoError(t, err)

	_, err = db.Conn.Exec(ctx, `
		insert into tenants (name, slug) values ($1, $2)
	`, "Acme 2", "acme-law")
	require.Error(t, err) // must fail due to unique constraint
}

func TestTenants_InsertAndSelect(t *testing.T) {
	db := testdb.NewPostgres(t)
	testdb.ApplyMigrations(t, db.Conn)

	ctx := context.Background()

	var id string
	err := db.Conn.QueryRow(ctx, `
		insert into tenants (name, slug) values ($1, $2)
		returning id::text
	`, "Acme", "acme-law").Scan(&id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	var gotName, gotSlug string
	err = db.Conn.QueryRow(ctx, `
		select name, slug from tenants where id = $1::uuid
	`, id).Scan(&gotName, &gotSlug)
	require.NoError(t, err)
	require.Equal(t, "Acme", gotName)
	require.Equal(t, "acme-law", gotSlug)
}
