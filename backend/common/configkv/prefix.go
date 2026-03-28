package configkv

import "strings"

// GlobalConfigPrefix is the root prefix for shared pay platform config in Consul KV (e.g. pay/config/global/).
func GlobalConfigPrefix() string {
	return "pay/config/global/"
}

// ServiceConfigPrefix is the per-service overlay prefix (e.g. pay/config/services/{name}/).
func ServiceConfigPrefix(serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "pay/config/services/unknown/"
	}
	return "pay/config/services/" + serviceName + "/"
}
