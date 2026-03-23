// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/openapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MerchantSignMiddleware struct {
	merchants merchantclient.Merchant
}

func NewMerchantSignMiddleware(merchants merchantclient.Merchant) *MerchantSignMiddleware {
	return &MerchantSignMiddleware{merchants: merchants}
}

func (m *MerchantSignMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := readParams(r)
		if err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_PARAMS", "invalid params")
			return
		}
		merchantId := params["merchant_id"]
		if merchantId == "" {
			openapi.Write(w, http.StatusBadRequest, "MERCHANT_ID_REQUIRED", "merchant_id required")
			return
		}
		sign := params["sign"]
		if sign == "" {
			openapi.Write(w, http.StatusBadRequest, "SIGN_REQUIRED", "sign required")
			return
		}

		auth, err := m.merchants.GetAuthInfo(r.Context(), &merchantclient.GetAuthInfoReq{MerchantId: merchantId})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				openapi.Write(w, http.StatusUnauthorized, "MERCHANT_NOT_FOUND", "merchant not found")
				return
			}
			openapi.Write(w, http.StatusInternalServerError, "INTERNAL_ERROR", "merchant lookup failed")
			return
		}
		if auth.GetStatus() != 1 {
			openapi.Write(w, http.StatusUnauthorized, "MERCHANT_DISABLED", "merchant disabled")
			return
		}
		if !ipAllowed(r.Context(), r.RemoteAddr, auth.GetIpWhitelist()) {
			openapi.Write(w, http.StatusForbidden, "IP_NOT_ALLOWED", "ip not allowed")
			return
		}

		expect := md5Sign(params, auth.GetApiSecret())
		if !strings.EqualFold(expect, sign) {
			openapi.Write(w, http.StatusUnauthorized, "INVALID_SIGN", "invalid sign")
			return
		}

		next(w, r)
	}
}

func readParams(r *http.Request) (map[string]string, error) {
	params := map[string]string{}
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			params[strings.ToLower(k)] = vs[0]
		}
	}

	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		var raw map[string]any
		if len(body) > 0 {
			if err := json.Unmarshal(body, &raw); err != nil {
				return nil, err
			}
		}
		for k, v := range raw {
			if v == nil {
				continue
			}
			params[strings.ToLower(k)] = anyToString(v)
		}
	}
	return params, nil
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func md5Sign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		k = strings.ToLower(k)
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		if i > 0 && b.Len() > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	if b.Len() > 0 {
		b.WriteByte('&')
	}
	b.WriteString("key=")
	b.WriteString(secret)

	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func ipAllowed(ctx context.Context, remoteAddr string, whitelist string) bool {
	whitelist = strings.TrimSpace(whitelist)
	if whitelist == "" {
		return true
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, item := range strings.Split(whitelist, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.Contains(item, "/") {
			_, cidr, err := net.ParseCIDR(item)
			if err == nil && cidr.Contains(ip) {
				return true
			}
			continue
		}
		if net.ParseIP(item) != nil && item == host {
			return true
		}
	}
	return false
}
