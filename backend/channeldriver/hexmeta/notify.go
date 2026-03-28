package hexmeta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gloopai/pay/channeldriver/base"
)

func verifyNotifyBody(body []byte, secret string) (map[string]string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var m map[string]interface{}
	if err := dec.Decode(&m); err != nil || m == nil {
		return nil, base.ErrVerifyNotify
	}
	sigRaw, ok := m["sign"]
	if !ok {
		return nil, base.ErrVerifyNotify
	}
	sigStr := strings.TrimSpace(valueForSign(sigRaw))
	if sigStr == "" {
		return nil, base.ErrVerifyNotify
	}
	params := make(map[string]string)
	for k, v := range m {
		if k == "sign" {
			continue
		}
		s := valueForSign(v)
		if s == "" {
			continue
		}
		params[k] = s
	}
	if strings.ToLower(strings.TrimSpace(sigStr)) != signMD5(params, secret) {
		return nil, base.ErrVerifyNotify
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case string:
			out[k] = strings.TrimSpace(t)
		case json.Number:
			out[k] = t.String()
		case nil:
			out[k] = ""
		default:
			out[k] = strings.TrimSpace(fmt.Sprint(t))
		}
	}
	return out, nil
}

func valueForSign(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.0f", t), "0"), ".")
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func readNotifyBody(r *http.Request) ([]byte, error) {
	if r == nil || r.Body == nil {
		return nil, base.ErrVerifyNotify
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, base.ErrVerifyNotify
	}
	return raw, nil
}
