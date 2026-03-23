package config

import "time"

type Config struct {
	Name     string
	Timezone string
	Nsq      struct {
		NsqdTCPAddr string
		Topic       string
		Channel     string
		MaxAttempts int `json:",optional"`
	}
	Mysql struct {
		DataSource string
	}
	Http struct {
		Timeout time.Duration `json:",optional"`
	}
	Consul struct {
		Addr    string
		Service string
		ID      string `json:",optional"`
		Host    string `json:",optional"`
	}
	Health struct {
		ListenOn string
	}
}
