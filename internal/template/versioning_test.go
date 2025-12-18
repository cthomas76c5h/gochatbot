package template_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/template"
)

func TestNextDraftVersion_Empty(t *testing.T) {
	v := template.NextDraftVersion(nil, "v1")
	require.Equal(t, 1, v.Number)
	require.Nil(t, v.PublishedAt)
	require.Equal(t, "v1", v.ID)
}

func TestNextDraftVersion_IncrementsMax(t *testing.T) {
	existing := []template.Version{
		{ID: "a", Number: 1},
		{ID: "b", Number: 3},
		{ID: "c", Number: 2},
	}
	v := template.NextDraftVersion(existing, "d")
	require.Equal(t, 4, v.Number)
	require.Nil(t, v.PublishedAt)
}

func TestPublishVersion_SetsPublishedAndUnpublishesOthers(t *testing.T) {
	t0 := time.Date(2025, 12, 18, 12, 0, 0, 0, time.UTC)

	existing := []template.Version{
		{ID: "a", Number: 1},
		{ID: "b", Number: 2},
		{ID: "c", Number: 3},
	}

	out, err := template.PublishVersion(existing, 2, t0)
	require.NoError(t, err)

	require.Nil(t, out[0].PublishedAt)
	require.NotNil(t, out[1].PublishedAt)
	require.Equal(t, t0, *out[1].PublishedAt)
	require.Nil(t, out[2].PublishedAt)
}

func TestPublishVersion_NotFound(t *testing.T) {
	t0 := time.Now()

	existing := []template.Version{{ID: "a", Number: 1}}
	_, err := template.PublishVersion(existing, 99, t0)
	require.ErrorIs(t, err, domain.ErrTemplateVersionNotFound)
}
