package pagination_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/pagination"
)

func TestCursor_EncodeDecode_RoundTrip(t *testing.T) {
	t0 := time.Date(2025, 12, 18, 12, 34, 56, 123456789, time.UTC)
	c0 := pagination.Cursor{CreatedAt: t0, ID: "m123"}

	enc := pagination.Encode(c0)
	got, err := pagination.Decode(enc)
	require.NoError(t, err)
	require.Equal(t, t0, got.CreatedAt)
	require.Equal(t, "m123", got.ID)
}

func TestCursor_Decode_InvalidBase64(t *testing.T) {
	_, err := pagination.Decode("!!!not-base64!!!")
	require.ErrorIs(t, err, domain.ErrInvalidCursor)
}

func TestCursor_Decode_MissingSeparator(t *testing.T) {
	// base64("no-sep-here") has no '|'
	s := "bm8tc2VwLWhlcmU"
	_, err := pagination.Decode(s)
	require.ErrorIs(t, err, domain.ErrInvalidCursor)
}

func TestCursor_Decode_InvalidTime(t *testing.T) {
	// base64("badtime|id")
	s := "YmFkdGltZXxpZA"
	_, err := pagination.Decode(s)
	require.ErrorIs(t, err, domain.ErrInvalidCursor)
}

func TestCursor_Decode_EmptyID(t *testing.T) {
	// base64("2025-12-18T00:00:00Z|")
	s := "MjAyNS0xMi0xOFQwMDowMDowMFp8"
	_, err := pagination.Decode(s)
	require.ErrorIs(t, err, domain.ErrInvalidCursor)
}
