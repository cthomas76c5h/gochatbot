package validate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/validate"
)

func TestNormalizeEmail_OK(t *testing.T) {
	got, err := validate.NormalizeEmail("  Chris.Thomas+tag@Example.COM  ")
	require.NoError(t, err)
	require.Equal(t, "chris.thomas+tag@example.com", got)
}

func TestNormalizeEmail_Invalid(t *testing.T) {
	_, err := validate.NormalizeEmail("not-an-email")
	require.ErrorIs(t, err, domain.ErrInvalidEmail)
}

func TestNormalizeEmail_Empty(t *testing.T) {
	_, err := validate.NormalizeEmail("   ")
	require.ErrorIs(t, err, domain.ErrInvalidEmail)
}
