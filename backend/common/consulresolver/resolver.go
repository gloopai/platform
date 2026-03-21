package consulresolver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/consul"
	"github.com/hashicorp/consul/api"
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
	client, err := consul.NewClient(consulAddr)
	if err != nil {
		return nil, err
	}
	r := &consulResolver{
		service: service,
		cc:      cc,
		client:  client,
		closeCh: make(chan struct{}),
	}
	go r.watch()
	return r, nil
}

type consulResolver struct {
	service string
	cc      resolver.ClientConn
	client  *api.Client
	closeCh chan struct{}
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
	if r.client == nil || r.service == "" {
		return lastIndex, fmt.Errorf("invalid consul target: %q", r.service)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 310*time.Second)
	defer cancel()

	opts := &api.QueryOptions{
		WaitIndex: lastIndex,
		WaitTime:  300 * time.Second,
	}
	entries, meta, err := r.client.Health().Service(strings.TrimPrefix(r.service, "/"), "", true, opts.WithContext(ctx))
	if err != nil {
		return lastIndex, err
	}

	addrs := make([]resolver.Address, 0, len(entries))
	for _, e := range entries {
		host := ""
		if e.Service != nil {
			host = strings.TrimSpace(e.Service.Address)
		}
		if host == "" {
			host = strings.TrimSpace(e.Node.Address)
		}
		if e.Service == nil || host == "" || e.Service.Port <= 0 {
			continue
		}
		addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", host, e.Service.Port)})
	}
	_ = r.cc.UpdateState(resolver.State{Addresses: addrs})

	if meta == nil || meta.LastIndex == 0 {
		return lastIndex, nil
	}
	return meta.LastIndex, nil
}
