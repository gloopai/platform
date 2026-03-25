package consulx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type ServiceNodeCheck struct {
	CheckID string `json:"check_id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Output  string `json:"output,omitempty"`
}

type ServiceNode struct {
	ServiceName string             `json:"service_name"`
	ServiceID   string             `json:"service_id"`
	Node        string             `json:"node"`
	Address     string             `json:"address"`
	Port        int                `json:"port"`
	Meta        map[string]string  `json:"meta,omitempty"`
	Checks      []ServiceNodeCheck `json:"checks,omitempty"`
}

func ListServiceNodes(ctx context.Context, consulAddr, service string) ([]ServiceNode, error) {
	service = strings.TrimSpace(service)
	if service == "" {
		return nil, fmt.Errorf("service required")
	}
	cli, err := NewClient(consulAddr)
	if err != nil {
		return nil, err
	}
	q := &api.QueryOptions{}
	if ctx != nil {
		if dl, ok := ctx.Deadline(); ok {
			q.WaitTime = time.Until(dl)
		}
	}
	entries, _, err := cli.Health().Service(service, "", false, q)
	if err != nil {
		return nil, err
	}
	out := make([]ServiceNode, 0, len(entries))
	for _, e := range entries {
		if e == nil || e.Service == nil || e.Node == nil {
			continue
		}
		n := ServiceNode{
			ServiceName: service,
			ServiceID:   e.Service.ID,
			Node:        e.Node.Node,
			Address:     e.Service.Address,
			Port:        e.Service.Port,
			Meta:        e.Service.Meta,
		}
		for _, c := range e.Checks {
			if c == nil {
				continue
			}
			n.Checks = append(n.Checks, ServiceNodeCheck{
				CheckID: c.CheckID,
				Name:    c.Name,
				Status:  c.Status,
				Output:  c.Output,
			})
		}
		out = append(out, n)
	}
	return out, nil
}
