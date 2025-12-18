package validate

import (
	"strings"

	"github.com/nyaruka/phonenumbers"

	"gochatbot/internal/domain"
)

func NormalizePhone(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", domain.ErrInvalidPhone
	}

	num, err := phonenumbers.Parse(s, "US")
	if err != nil {
		return "", domain.ErrInvalidPhone
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", domain.ErrInvalidPhone
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}
