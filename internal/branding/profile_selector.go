package branding

import "gochatbot/internal/domain"

// SelectEmailProfile chooses an email profile key to use.
// Rules:
// - If requested != "" and exists in profiles: use it
// - Else if defaultKey exists in profiles: use it
// - Else if profiles has exactly 1 entry: use that
// - Else error
func SelectEmailProfile(
	profiles map[string]struct{},
	defaultKey string,
	requested string,
) (string, error) {

	if requested != "" {
		if _, ok := profiles[requested]; ok {
			return requested, nil
		}
		return "", domain.ErrUnknownEmailProfile
	}

	if defaultKey != "" {
		if _, ok := profiles[defaultKey]; ok {
			return defaultKey, nil
		}
	}

	if len(profiles) == 1 {
		for k := range profiles {
			return k, nil
		}
	}

	return "", domain.ErrUnknownEmailProfile
}
