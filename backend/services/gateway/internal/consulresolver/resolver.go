package consulresolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/resolver"
)

var once sync.Once

func Register() {
	once.Do(func() {
		resolver.Register(&builder{})
	})
}

type builder struct{}

func (b *builder) Scheme() string { return "consul" }

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	consulAddr := strings.TrimSpace(target.URL.Host)
	service := strings.TrimPrefix(strings.TrimSpace(target.URL.Path), "/")
	if service == "" {
		service = strings.TrimSpace(target.Endpoint())
	}
	r := &consulResolver{
		consulAddr: consulAddr,
		service:    service,
		cc:         cc,
		client:     &http.Client{Timeout: 3 * time.Second},
		closeCh:    make(chan struct{}),
	}
	go r.watch()
	return r, nil
}

type consulResolver struct {
	consulAddr string
	service    string
	cc         resolver.ClientConn
	client     *http.Client
	closeCh    chan struct{}
}

func (r *consulResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *consulResolver) Close() { close(r.closeCh) }

func (r *consulResolver) watch() {
	_ = r.resolve()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = r.resolve()
		case <-r.closeCh:
			return
		}
	}
}

func (r *consulResolver) resolve() error {
	if r.consulAddr == "" || r.service == "" {
		return fmt.Errorf("invalid consul target: %q %q", r.consulAddr, r.service)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+r.consulAddr+"/v1/health/service/"+r.service+"?passing=true", nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("consul query failed: %s", resp.Status)
	}

	var entries []struct {
		Node struct {
			Address string `json:"Address"`
		} `json:"Node"`
		Service struct {
			Address string `json:"Address"`
			Port    int    `json:"Port"`
		} `json:"Service"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return err
	}

	addrs := make([]resolver.Address, 0, len(entries))
	for _, e := range entries {
		host := strings.TrimSpace(e.Service.Address)
		if host == "" {
			host = strings.TrimSpace(e.Node.Address)
		}
		if host == "" || e.Service.Port <= 0 {
			continue
		}
		addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", host, e.Service.Port)})
	}
	return r.cc.UpdateState(resolver.State{Addresses: addrs})
}

