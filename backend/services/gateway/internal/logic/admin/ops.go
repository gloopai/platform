package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
)

type AdminOps struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOps(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOps {
	return &AdminOps{ctx: ctx, svcCtx: svcCtx}
}

func (a *AdminOps) ServicesStatus() (*types.OpsServicesResp, error) {
	c := a.svcCtx.Config

	services := []string{}
	if s := strings.TrimSpace(c.Consul.Service); s != "" {
		services = append(services, s)
	}
	// zrpc target: consul://host:8500/serviceName
	if t := strings.TrimSpace(c.TradeRpc.Target); t != "" {
		if parts := strings.Split(t, "/"); len(parts) > 0 {
			services = append(services, parts[len(parts)-1])
		}
	}
	if t := strings.TrimSpace(c.CoreRpc.Target); t != "" {
		if parts := strings.Split(t, "/"); len(parts) > 0 {
			services = append(services, parts[len(parts)-1])
		}
	}
	// default include notice-consumer; can be overridden/extended by OpsMonitor.Services
	services = append(services, "payment.worker.notice-consumer")
	for _, s := range c.OpsMonitor.Services {
		services = append(services, strings.TrimSpace(s))
	}

	// de-dupe
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

	resp := &types.OpsServicesResp{Ok: true}
	for _, name := range uniq {
		nodes, err := consulx.ListServiceNodes(a.ctx, c.Consul.Addr, name)
		if err != nil {
			// if consul query fails, mark overall not ok but still return what we have
			resp.Ok = false
			resp.Services = append(resp.Services, types.OpsServiceStatus{
				ServiceName: name,
				Ok:          false,
				Nodes:       nil,
			})
			continue
		}
		st := summarizeService(name, nodes)
		if !st.Ok {
			resp.Ok = false
		}
		resp.Services = append(resp.Services, st)
	}
	return resp, nil
}

func summarizeService(service string, nodes []consulx.ServiceNode) types.OpsServiceStatus {
	out := types.OpsServiceStatus{ServiceName: service, Ok: true}
	out.Total = len(nodes)
	for _, n := range nodes {
		status := worstCheckStatus(n.Checks)
		out.Nodes = append(out.Nodes, types.OpsServiceNode{
			ServiceName: n.ServiceName,
			ServiceID:   n.ServiceID,
			Node:        n.Node,
			Address:     n.Address,
			Port:        n.Port,
			Status:      status,
		})
		switch status {
		case "passing":
			out.Passing++
		case "warning":
			out.Warning++
			out.Ok = false
		case "critical":
			out.Critical++
			out.Ok = false
		default:
			out.Ok = false
		}
	}
	if out.Total == 0 {
		out.Ok = false
	}
	return out
}

func worstCheckStatus(checks []consulx.ServiceNodeCheck) string {
	// Consul status: passing / warning / critical
	// We'll treat unknown as critical-ish.
	worst := "passing"
	for _, c := range checks {
		switch c.Status {
		case "critical":
			return "critical"
		case "warning":
			worst = "warning"
		case "passing":
		default:
			worst = "unknown"
		}
	}
	return worst
}
