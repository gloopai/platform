package hexmeta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const basePath = "/exposed/v1"

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (d *Driver) postJSON(ctx context.Context, path string, body map[string]string) (json.RawMessage, error) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := strings.TrimRight(d.cfg.GatewayURL, "/") + basePath + path
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())
	body["appId"] = d.cfg.AppID
	body["timestamp"] = ts
	body["sign"] = signMD5(body, d.cfg.Secret)

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hexmeta: HTTP %d: %s", resp.StatusCode, truncate(string(raw), 512))
	}
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("hexmeta: decode envelope: %w", err)
	}
	if env.Code != 1 {
		return nil, fmt.Errorf("hexmeta: api code=%d msg=%s", env.Code, env.Msg)
	}
	return env.Data, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
