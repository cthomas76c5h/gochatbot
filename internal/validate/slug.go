package validate

import (
	"strings"
	"unicode"

	"gochatbot/internal/domain"
)

// NormalizeSlug converts arbitrary input to a URL-safe slug.
// Rules:
// - lowercase
// - trim whitespace
// - letters/digits kept
// - whitespace/_ converted to '-'
// - all other chars removed
// - collapse multiple '-' into one
// - trim leading/trailing '-'
// - length must be 3..63
func NormalizeSlug(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", domain.ErrInvalidSlug
	}

	var b strings.Builder
	b.Grow(len(s))

	lastWasDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastWasDash = false

		case unicode.IsSpace(r) || r == '_' || r == '-':
			if !lastWasDash && b.Len() > 0 {
				b.WriteByte('-')
				lastWasDash = true
			}

		default:
			// drop punctuation/symbols
		}
	}

	out := b.String()
	out = strings.Trim(out, "-")

	if len(out) < 3 || len(out) > 63 {
		return "", domain.ErrInvalidSlug
	}
	return out, nil
}
