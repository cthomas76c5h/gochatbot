package validate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/validate"
)

func TestNormalizeSlug_OK(t *testing.T) {
	got, err := validate.NormalizeSlug("  Acme Law Firm!! ")
	require.NoError(t, err)
	require.Equal(t, "acme-law-firm", got)
}

func TestNormalizeSlug_CollapseHyphens(t *testing.T) {
	got, err := validate.NormalizeSlug("Acme---Law___Firm")
	require.NoError(t, err)
	require.Equal(t, "acme-law-firm", got)
}

func TestNormalizeSlug_TrimsDashes(t *testing.T) {
	got, err := validate.NormalizeSlug(" --- Acme --- ")
	require.NoError(t, err)
	require.Equal(t, "acme", got)
}

func TestNormalizeSlug_RemovesPunctuation(t *testing.T) {
	got, err := validate.NormalizeSlug("A.C.M.E.!!!")
	require.NoError(t, err)
	require.Equal(t, "acme", got)
}

func TestNormalizeSlug_TooShort(t *testing.T) {
	_, err := validate.NormalizeSlug("a")
	require.ErrorIs(t, err, domain.ErrInvalidSlug)
}

func TestNormalizeSlug_Empty(t *testing.T) {
	_, err := validate.NormalizeSlug("   ")
	require.ErrorIs(t, err, domain.ErrInvalidSlug)
}

func TestNormalizeSlug_TooLong(t *testing.T) {
	// 64 chars after normalization should fail
	_, err := validate.NormalizeSlug("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.ErrorIs(t, err, domain.ErrInvalidSlug)
}
