package testdb

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type DB struct {
	Conn *pgx.Conn
}

func NewPostgres(t *testing.T) *DB {
	t.Helper()
	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
	)
	require.NoError(t, err)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	var conn *pgx.Conn
	var lastErr error

	// Postgres may not accept connections immediately even if container is "ready".
	for range 40 { // ~10s total
		conn, lastErr = pgx.Connect(ctx, connStr)
		if lastErr == nil {
			// also verify the server responds
			if pingErr := conn.Ping(ctx); pingErr == nil {
				lastErr = nil
				break
			} else {
				_ = conn.Close(ctx)
				lastErr = pingErr
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	require.NoError(t, lastErr)

	t.Cleanup(func() {
		_ = conn.Close(ctx)
		_ = pg.Terminate(ctx)
	})

	return &DB{Conn: conn}
}

func ApplyMigrations(t *testing.T, conn *pgx.Conn) {
	t.Helper()
	ctx := context.Background()

	sql := readAllMigrations(t)
	_, err := conn.Exec(ctx, sql)
	require.NoError(t, err)
}

func readAllMigrations(t *testing.T) string {
	t.Helper()

	// find repo root by walking up until we see /migrations
	dir, err := os.Getwd()
	require.NoError(t, err)

	var migDir string
	for i := 0; i < 10; i++ {
		candidate := filepath.Join(dir, "migrations")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			migDir = candidate
			break
		}
		dir = filepath.Dir(dir)
	}
	require.NotEmpty(t, migDir, "could not find migrations directory")

	entries, err := os.ReadDir(migDir)
	require.NoError(t, err)

	// read in filename order (0001_, 0002_, etc.)
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	require.NotEmpty(t, names, "no .sql migrations found")

	sortStrings(names)

	var b strings.Builder
	for _, name := range names {
		path := filepath.Join(migDir, name)
		body, err := os.ReadFile(path)
		require.NoError(t, err)
		b.WriteString("\n-- ")
		b.WriteString(name)
		b.WriteString("\n")
		b.Write(body)
		b.WriteString("\n")
	}

	return b.String()
}

// tiny lexical sort (0001_ prefix means lexical == numeric order)
func sortStrings(a []string) {
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if a[j] < a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}
