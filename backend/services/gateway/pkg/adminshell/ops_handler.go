package adminshell

import (
	"net/http"
	"strings"

	"github.com/gloopai/platform/gateway/internal/apiresp"
)

// OpsServicesHandler returns GET /v1/admin/ops/services：按 ShellConfig 组装 Consul 服务名，经 ServiceHub（core 上）查询实例状态。
func OpsServicesHandler(o ShellOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := opsMonitorServiceNames(o.Config)
		resp, err := o.ServiceHub.GetOpsServicesStatus(r.Context(), names)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
			return
		}
		apiresp.OK(w, resp)
	}
}

func opsMonitorServiceNames(c ShellConfig) []string {
	var services []string
	if s := strings.TrimSpace(c.Consul.Service); s != "" {
		services = append(services, s)
	}
	if t := strings.TrimSpace(c.ServiceHubRpc.Target); t != "" {
		if parts := strings.Split(t, "/"); len(parts) > 0 {
			services = append(services, parts[len(parts)-1])
		}
	}
	for _, s := range c.OpsMonitor.Services {
		services = append(services, strings.TrimSpace(s))
	}
	seen := map[string]struct{}{}
	var uniq []string
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		uniq = append(uniq, s)
	}
	return uniq
}
