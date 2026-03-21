package consulresolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
		client:     &http.Client{},
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
	var (
		lastIndex uint64
		backoff   = 200 * time.Millisecond
	)
	for {
		select {
		case <-r.closeCh:
			return
		default:
		}

		nextIndex, err := r.resolve(lastIndex)
		if err != nil {
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
			case <-r.closeCh:
				timer.Stop()
				return
			}
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = 200 * time.Millisecond
		if nextIndex > lastIndex {
			lastIndex = nextIndex
		}
	}
}

func (r *consulResolver) resolve(lastIndex uint64) (uint64, error) {
	if r.consulAddr == "" || r.service == "" {
		return lastIndex, fmt.Errorf("invalid consul target: %q %q", r.consulAddr, r.service)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 310*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://%s/v1/health/service/%s?passing=true&wait=300s&index=%d",
		r.consulAddr,
		strings.TrimPrefix(r.service, "/"),
		lastIndex,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return lastIndex, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return lastIndex, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return lastIndex, fmt.Errorf("consul query failed: %s", resp.Status)
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
		return lastIndex, err
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
	_ = r.cc.UpdateState(resolver.State{Addresses: addrs})

	headerIndex := strings.TrimSpace(resp.Header.Get("X-Consul-Index"))
	if headerIndex == "" {
		return lastIndex, nil
	}
	v, err := strconv.ParseUint(headerIndex, 10, 64)
	if err != nil {
		return lastIndex, nil
	}
	return v, nil
}
