package branding_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/branding"
	"gochatbot/internal/domain"
)

func set(keys ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

func TestSelectEmailProfile_RequestedWins(t *testing.T) {
	got, err := branding.SelectEmailProfile(
		set("answering_legal", "ring_savvy"),
		"answering_legal",
		"ring_savvy",
	)
	require.NoError(t, err)
	require.Equal(t, "ring_savvy", got)
}

func TestSelectEmailProfile_RequestedMissingFails(t *testing.T) {
	_, err := branding.SelectEmailProfile(
		set("answering_legal"),
		"answering_legal",
		"ring_savvy",
	)
	require.ErrorIs(t, err, domain.ErrUnknownEmailProfile)
}

func TestSelectEmailProfile_DefaultUsedWhenNoRequested(t *testing.T) {
	got, err := branding.SelectEmailProfile(
		set("answering_legal", "ring_savvy"),
		"answering_legal",
		"",
	)
	require.NoError(t, err)
	require.Equal(t, "answering_legal", got)
}

func TestSelectEmailProfile_SingleFallback(t *testing.T) {
	got, err := branding.SelectEmailProfile(
		set("only_profile"),
		"",
		"",
	)
	require.NoError(t, err)
	require.Equal(t, "only_profile", got)
}

func TestSelectEmailProfile_NoDefaultMultipleProfilesFails(t *testing.T) {
	_, err := branding.SelectEmailProfile(
		set("a", "b"),
		"",
		"",
	)
	require.ErrorIs(t, err, domain.ErrUnknownEmailProfile)
}
