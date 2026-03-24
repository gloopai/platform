// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/pay/common/grpcclient/merchantclient"
	"github.com/gloopai/pay/gateway/internal/openapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MerchantSignMiddleware struct {
	merchants            merchantclient.Merchant
	replayGuard          ReplayGuard
	allowedSkewSeconds   int64
	trustForwardedForIPs bool
}

func NewMerchantSignMiddleware(merchants merchantclient.Merchant, replayGuard ReplayGuard, allowedSkewSeconds int64, trustForwardedForIPs bool) *MerchantSignMiddleware {
	if allowedSkewSeconds <= 0 {
		allowedSkewSeconds = 300
	}
	return &MerchantSignMiddleware{
		merchants:            merchants,
		replayGuard:          replayGuard,
		allowedSkewSeconds:   allowedSkewSeconds,
		trustForwardedForIPs: trustForwardedForIPs,
	}
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
		tsRaw := strings.TrimSpace(params["timestamp"])
		if tsRaw == "" {
			openapi.Write(w, http.StatusBadRequest, "TIMESTAMP_REQUIRED", "timestamp required")
			return
		}
		ts, err := strconv.ParseInt(tsRaw, 10, 64)
		if err != nil {
			openapi.Write(w, http.StatusBadRequest, "INVALID_TIMESTAMP", "invalid timestamp")
			return
		}
		now := time.Now().Unix()
		if ts < now-m.allowedSkewSeconds || ts > now+m.allowedSkewSeconds {
			openapi.Write(w, http.StatusBadRequest, "INVALID_TIMESTAMP", "timestamp out of allowed window")
			return
		}
		nonce := strings.TrimSpace(params["nonce"])
		if nonce == "" {
			openapi.Write(w, http.StatusBadRequest, "NONCE_REQUIRED", "nonce required")
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
		clientHost := ClientHost(r, m.trustForwardedForIPs)
		if !ipAllowed(clientHost, auth.GetIpWhitelist()) {
			openapi.Write(w, http.StatusForbidden, "IP_NOT_ALLOWED", "ip not allowed")
			return
		}

		expect := Md5Sign(params, auth.GetApiSecret())
		if !strings.EqualFold(expect, sign) {
			openapi.Write(w, http.StatusUnauthorized, "INVALID_SIGN", "invalid sign")
			return
		}
		if m.replayGuard != nil {
			ok, err := m.replayGuard.MarkSeen(r.Context(), merchantId, nonce, ts)
			if err != nil {
				openapi.Write(w, http.StatusServiceUnavailable, "UNAVAILABLE", "replay guard unavailable")
				return
			}
			if !ok {
				openapi.Write(w, http.StatusConflict, "REPLAY_REQUEST", "replay request rejected")
				return
			}
		}

		next(w, r)
	}
}

func ipAllowed(clientHost string, whitelist string) bool {
	whitelist = strings.TrimSpace(whitelist)
	if whitelist == "" {
		return true
	}
	host := strings.TrimSpace(clientHost)
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
