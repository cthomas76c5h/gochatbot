package validate

import (
	"strings"

	"github.com/asaskevich/govalidator"

	"gochatbot/internal/domain"
)

func NormalizeEmail(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", domain.ErrInvalidEmail
	}
	s = strings.ToLower(s)

	if !govalidator.IsEmail(s) {
		return "", domain.ErrInvalidEmail
	}
	return s, nil
}
