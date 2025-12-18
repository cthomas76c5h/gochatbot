package validate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/validate"
)

func TestNormalizeHexColor_OK(t *testing.T) {
	got, err := validate.NormalizeHexColor("  #Aa11Ff ")
	require.NoError(t, err)
	require.Equal(t, "#aa11ff", got)
}

func TestNormalizeHexColor_OK_NoHash(t *testing.T) {
	got, err := validate.NormalizeHexColor("Aa11Ff")
	require.NoError(t, err)
	require.Equal(t, "#aa11ff", got)
}

func TestNormalizeHexColor_Invalid(t *testing.T) {
	_, err := validate.NormalizeHexColor("#gg11ff")
	require.ErrorIs(t, err, domain.ErrInvalidColor)
}

func TestNormalizeHexColor_InvalidLength(t *testing.T) {
	_, err := validate.NormalizeHexColor("#12345")
	require.ErrorIs(t, err, domain.ErrInvalidColor)
}
