package repo

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// 23505 = unique_violation
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.TrimSpace(pgErr.Code) == "23505"
	}
	return false
}
