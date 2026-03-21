package consul

import (
	"net/url"
	"strings"
	"sync"

	"github.com/hashicorp/consul/api"
)

type BaseConfig struct {
	Addr       string
	Token      string
	Datacenter string
	Namespace  string
}

var (
	baseMu     sync.RWMutex
	baseConfig BaseConfig

	clientMu    sync.Mutex
	clientCache = map[string]*api.Client{}
)

func SetBaseConfig(cfg BaseConfig) {
	baseMu.Lock()
	baseConfig = cfg
	baseMu.Unlock()
}

func NewClient(consulAddr string) (*api.Client, error) {
	addr := strings.TrimSpace(consulAddr)
	if addr == "" {
		baseMu.RLock()
		addr = strings.TrimSpace(baseConfig.Addr)
		baseMu.RUnlock()
	}

	key := normalizeAddrKey(addr)
	clientMu.Lock()
	if cli, ok := clientCache[key]; ok {
		clientMu.Unlock()
		return cli, nil
	}
	clientMu.Unlock()

	cfg := api.DefaultConfig()
	baseMu.RLock()
	bc := baseConfig
	baseMu.RUnlock()
	if bc.Token != "" {
		cfg.Token = bc.Token
	}
	if bc.Datacenter != "" {
		cfg.Datacenter = bc.Datacenter
	}
	if bc.Namespace != "" {
		cfg.Namespace = bc.Namespace
	}

	if addr != "" {
		if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
			u, err := url.Parse(addr)
			if err != nil {
				return nil, err
			}
			if u.Scheme != "" {
				cfg.Scheme = u.Scheme
			}
			cfg.Address = u.Host
		} else {
			cfg.Address = addr
		}
	}

	cli, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	clientMu.Lock()
	clientCache[key] = cli
	clientMu.Unlock()
	return cli, nil
}

func normalizeAddrKey(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	return strings.ToLower(addr)
}
