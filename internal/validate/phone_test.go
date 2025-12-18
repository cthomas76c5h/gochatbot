package validate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/validate"
)

func TestNormalizePhone_US_OK(t *testing.T) {
	got, err := validate.NormalizePhone("(917) 555-1234")
	require.NoError(t, err)
	require.Equal(t, "+19175551234", got)
}

func TestNormalizePhone_AlreadyE164_OK(t *testing.T) {
	got, err := validate.NormalizePhone("+14155552671")
	require.NoError(t, err)
	require.Equal(t, "+14155552671", got)
}

func TestNormalizePhone_Invalid(t *testing.T) {
	_, err := validate.NormalizePhone("abc lol nope")
	require.ErrorIs(t, err, domain.ErrInvalidPhone)
}
