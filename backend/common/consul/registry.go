package consul

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/hashicorp/consul/api"
)

type Registrar struct {
	serviceID string
	client    *api.Client
}

func Register(consulAddr, serviceName, serviceID, listenOn, host string) (*Registrar, error) {
	consulAddr = strings.TrimSpace(consulAddr)
	if consulAddr == "" {
		return nil, errors.New("consul addr required")
	}
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, errors.New("consul service name required")
	}

	lh, lp, err := net.SplitHostPort(listenOn)
	if err != nil {
		return nil, err
	}
	if host == "" || host == "0.0.0.0" {
		if lh != "" && lh != "0.0.0.0" {
			host = lh
		} else {
			host = "127.0.0.1"
		}
	}
	port, err := parsePort(lp)
	if err != nil {
		return nil, err
	}
	if serviceID == "" {
		serviceID = fmt.Sprintf("%s-%s-%d", serviceName, host, port)
	}

	checkHost := host
	if host == "127.0.0.1" || host == "localhost" {
		client, err := newClient(consulAddr)
		if err != nil {
			return nil, err
		}
		nodeName := consulNodeName(client)
		if isLikelyDockerNodeName(nodeName) {
			checkHost = "host.docker.internal"
		}
		reg := &api.AgentServiceRegistration{
			Name:    serviceName,
			ID:      serviceID,
			Address: host,
			Port:    port,
			Check: &api.AgentServiceCheck{
				TCP:                            fmt.Sprintf("%s:%d", checkHost, port),
				Interval:                       "10s",
				DeregisterCriticalServiceAfter: "1m",
			},
		}
		if err := client.Agent().ServiceRegister(reg); err != nil {
			return nil, err
		}
		return &Registrar{serviceID: serviceID, client: client}, nil
	}

	client, err := newClient(consulAddr)
	if err != nil {
		return nil, err
	}
	reg := &api.AgentServiceRegistration{
		Name:    serviceName,
		ID:      serviceID,
		Address: host,
		Port:    port,
		Check: &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", checkHost, port),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}
	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, err
	}

	return &Registrar{serviceID: serviceID, client: client}, nil
}

func (r *Registrar) Deregister() error {
	if r == nil || r.serviceID == "" || r.client == nil {
		return nil
	}
	return r.client.Agent().ServiceDeregister(r.serviceID)
}

func parsePort(s string) (int, error) {
	var p int
	_, err := fmt.Sscanf(s, "%d", &p)
	if err != nil {
		return 0, err
	}
	if p <= 0 || p > 65535 {
		return 0, fmt.Errorf("invalid port: %d", p)
	}
	return p, nil
}

func newClient(consulAddr string) (*api.Client, error) {
	cfg := api.DefaultConfig()
	consulAddr = strings.TrimSpace(consulAddr)
	if consulAddr != "" {
		if strings.HasPrefix(consulAddr, "http://") || strings.HasPrefix(consulAddr, "https://") {
			u, err := url.Parse(consulAddr)
			if err != nil {
				return nil, err
			}
			if u.Scheme != "" {
				cfg.Scheme = u.Scheme
			}
			cfg.Address = u.Host
		} else {
			cfg.Address = consulAddr
		}
	}
	return api.NewClient(cfg)
}

func consulNodeName(client *api.Client) string {
	if client == nil {
		return ""
	}
	self, err := client.Agent().Self()
	if err != nil {
		return ""
	}
	cfg, ok := self["Config"]
	if !ok {
		return ""
	}
	nodeName, _ := cfg["NodeName"].(string)
	return strings.TrimSpace(nodeName)
}

func isLikelyDockerNodeName(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) != 12 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
