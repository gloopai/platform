package gatewaymw

import "strings"

// MatchPathPattern matches a request path against a pattern with ":param" segments
// (same semantics as admin_api_rules.path_pattern).
func MatchPathPattern(pattern, path string) bool {
	if pattern == path {
		return true
	}
	ps := splitPathSegments(pattern)
	as := splitPathSegments(path)
	if len(ps) != len(as) {
		return false
	}
	for i := 0; i < len(ps); i++ {
		if strings.HasPrefix(ps[i], ":") {
			if as[i] == "" {
				return false
			}
			continue
		}
		if ps[i] != as[i] {
			return false
		}
	}
	return true
}

func splitPathSegments(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "/")
}
