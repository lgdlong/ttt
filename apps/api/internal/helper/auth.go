package helper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseDurationWithDays parses duration strings that support "d" suffix for days
// Supported formats: "1d", "24h", "3600s", "720h", etc.
// Since Go's time.ParseDuration doesn't support days, we handle it specially
func ParseDurationWithDays(s string) (time.Duration, error) {
	// Check if string ends with "d" for days
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.ParseInt(daysStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid days format: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	// Otherwise use standard Go duration parsing
	return time.ParseDuration(s)
}
