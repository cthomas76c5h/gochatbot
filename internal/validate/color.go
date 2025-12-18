package validate

import (
	"strings"

	"gochatbot/internal/domain"
)

func NormalizeHexColor(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "", domain.ErrInvalidColor
	}

	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return "", domain.ErrInvalidColor
	}

	for _, c := range s {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f')) {
			return "", domain.ErrInvalidColor
		}
	}

	return "#" + s, nil
}
