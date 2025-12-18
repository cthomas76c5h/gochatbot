package template

import (
	"sort"
	"time"

	"gochatbot/internal/domain"
)

type Version struct {
	ID          string // keep string for now; later uuid.UUID
	Number      int
	PublishedAt *time.Time
}

// NextDraftVersion returns a new draft version with Number = max(existing)+1.
func NextDraftVersion(existing []Version, newID string) Version {
	max := 0
	for _, v := range existing {
		if v.Number > max {
			max = v.Number
		}
	}
	return Version{
		ID:     newID,
		Number: max + 1,
		// PublishedAt nil => draft
	}
}

// PublishVersion marks the given version as published and unpublishes others.
func PublishVersion(existing []Version, versionNumber int, now time.Time) ([]Version, error) {
	found := false
	out := make([]Version, len(existing))
	copy(out, existing)

	for i := range out {
		if out[i].Number == versionNumber {
			found = true
			out[i].PublishedAt = ptrTime(now)
		} else {
			out[i].PublishedAt = nil
		}
	}

	if !found {
		return nil, domain.ErrTemplateVersionNotFound
	}

	// keep stable ordering by Number ascending
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func ptrTime(t time.Time) *time.Time { return &t }
