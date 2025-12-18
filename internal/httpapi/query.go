package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

func parseLimit(r *http.Request, def, min, max int) (int, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return def, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if n < min {
		n = min
	}
	if n > max {
		n = max
	}
	return n, true
}
