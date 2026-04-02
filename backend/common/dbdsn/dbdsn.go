package dbdsn

import (
	"net/url"
	"strings"

	"github.com/gloopai/platform/common/timex"
)

// WithTimezone ensures MySQL DSN uses local parse location and a configured session time_zone.
func WithTimezone(dataSource, timezone string) string {
	tz := timex.Resolve(timezone)

	parts := strings.SplitN(dataSource, "?", 2)
	if len(parts) == 1 {
		return parts[0] + "?loc=Local&time_zone=%27" + url.QueryEscape(tz) + "%27"
	}

	values, err := url.ParseQuery(parts[1])
	if err != nil {
		sep := "&"
		if strings.TrimSpace(parts[1]) == "" {
			sep = ""
		}
		return dataSource + sep + "loc=Local&time_zone=%27" + url.QueryEscape(tz) + "%27"
	}

	values.Set("loc", "Local")
	values.Set("time_zone", "'"+tz+"'")
	return parts[0] + "?" + values.Encode()
}
