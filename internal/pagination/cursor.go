package pagination

import (
	"encoding/base64"
	"strings"
	"time"

	"gochatbot/internal/domain"
)

type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// Encode cursor as base64("RFC3339Nano|id")
func Encode(c Cursor) string {
	raw := c.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + c.ID
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func Decode(s string) (Cursor, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Cursor{}, domain.ErrInvalidCursor
	}

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, domain.ErrInvalidCursor
	}

	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return Cursor{}, domain.ErrInvalidCursor
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return Cursor{}, domain.ErrInvalidCursor
	}

	id := strings.TrimSpace(parts[1])
	if id == "" {
		return Cursor{}, domain.ErrInvalidCursor
	}

	return Cursor{CreatedAt: t, ID: id}, nil
}
