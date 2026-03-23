package timex

import (
	"fmt"
	"strings"
	"time"
)

const DefaultTimezone = "Asia/Shanghai"

func Resolve(timezone string) string {
	tz := strings.TrimSpace(timezone)
	if tz == "" {
		return DefaultTimezone
	}
	return tz
}

func ApplyProcessTimezone(timezone string) error {
	tz := Resolve(timezone)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	time.Local = loc
	return nil
}
