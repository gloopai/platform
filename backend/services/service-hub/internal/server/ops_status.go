package server

import (
	"context"
	"strings"

	"github.com/gloopai/platform/common/consulx"
	"github.com/gloopai/platform/common/pb/servicehub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServiceHubServer) GetOpsServicesStatus(ctx context.Context, req *servicehub.GetOpsServicesStatusReq) (*servicehub.GetOpsServicesStatusResp, error) {
	addr := strings.TrimSpace(s.svcCtx.Config.Consul.Addr)
	if addr == "" {
		return nil, status.Error(codes.FailedPrecondition, "consul addr not configured")
	}
	names := dedupeOpsServiceNames(req.GetServiceNames())
	if len(names) == 0 {
		return &servicehub.GetOpsServicesStatusResp{Ok: true}, nil
	}
	resp := &servicehub.GetOpsServicesStatusResp{Ok: true}
	for _, name := range names {
		nodes, err := consulx.ListServiceNodes(ctx, addr, name)
		if err != nil {
			resp.Ok = false
			resp.Services = append(resp.Services, &servicehub.OpsSvcStatus{
				ServiceName: name,
				Ok:          false,
			})
			continue
		}
		st := summarizeOpsSvcStatus(name, nodes)
		if !st.GetOk() {
			resp.Ok = false
		}
		resp.Services = append(resp.Services, st)
	}
	return resp, nil
}

func dedupeOpsServiceNames(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func summarizeOpsSvcStatus(service string, nodes []consulx.ServiceNode) *servicehub.OpsSvcStatus {
	out := &servicehub.OpsSvcStatus{ServiceName: service, Ok: true}
	out.Total = int32(len(nodes))
	for _, n := range nodes {
		st := worstConsulCheckStatus(n.Checks)
		out.Nodes = append(out.Nodes, &servicehub.OpsSvcNode{
			ServiceName: n.ServiceName,
			ServiceId:   n.ServiceID,
			Node:        n.Node,
			Address:     n.Address,
			Port:        int32(n.Port),
			Status:      st,
		})
		switch st {
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

func worstConsulCheckStatus(checks []consulx.ServiceNodeCheck) string {
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
