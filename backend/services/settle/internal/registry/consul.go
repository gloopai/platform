package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type Registrar struct {
	consulAddr string
	serviceID  string
	client     *http.Client
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

	payload := map[string]any{
		"Name":    serviceName,
		"ID":      serviceID,
		"Address": host,
		"Port":    port,
		"Check": map[string]any{
			"TCP":                           fmt.Sprintf("%s:%d", host, port),
			"Interval":                      "10s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodPut, "http://"+consulAddr+"/v1/agent/service/register", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("consul register failed: %s", resp.Status)
	}

	return &Registrar{
		consulAddr: consulAddr,
		serviceID:  serviceID,
		client:     client,
	}, nil
}

func (r *Registrar) Deregister() error {
	if r == nil || r.serviceID == "" || r.consulAddr == "" {
		return nil
	}
	req, err := http.NewRequest(http.MethodPut, "http://"+r.consulAddr+"/v1/agent/service/deregister/"+r.serviceID, nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
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
