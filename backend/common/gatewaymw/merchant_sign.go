package gatewaymw

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/platform/common/signmd5"
)

// ErrMerchantSignNotFound is returned by [MerchantOpenAPISignLookup] when app_id is unknown.
var ErrMerchantSignNotFound = errors.New("merchant sign lookup: not found")

// MerchantOpenAPISignLookup resolves app_id to signing material for OpenAPI merchant requests.
type MerchantOpenAPISignLookup interface {
	LookupByAppID(ctx context.Context, appID string) (MerchantSignAccount, error)
}

// MerchantSignAccount is the data needed to verify OpenAPI signatures for one app.
type MerchantSignAccount struct {
	MerchantID  string
	AppSecret     string
	IPWhitelist string
	Active      bool
}

// MerchantSignCodes are business error codes (e.g. apiresp constants) for OpenAPI merchant sign failures.
type MerchantSignCodes struct {
	InvalidParams     int
	AppIDRequired     int
	SignRequired      int
	TimestampRequired int
	InvalidTimestamp  int
	NonceRequired     int
	MerchantNotFound  int
	MerchantDisabled  int
	IPNotAllowed      int
	InvalidSign       int
	Internal          int
	Unavailable       int
	ReplayRequest     int
}

// MerchantSignOptions configures [MerchantSign] (OpenAPI HMAC/MD5 style verification).
type MerchantSignOptions struct {
	Lookup               MerchantOpenAPISignLookup
	ReplayGuard          ReplayGuard
	AllowedSkewSeconds   int64
	TrustForwardedForIPs bool
	Fail                 func(w http.ResponseWriter, code int, message string)
	Codes                MerchantSignCodes
}

// MerchantSign verifies app_id / sign / timestamp / nonce and optional IP whitelist + replay guard.
type MerchantSign struct {
	lookup               MerchantOpenAPISignLookup
	replayGuard          ReplayGuard
	allowedSkewSeconds   int64
	trustForwardedForIPs bool
	fail                 func(w http.ResponseWriter, code int, message string)
	codes                MerchantSignCodes
}

// NewMerchantSign builds OpenAPI merchant signature middleware. Allowed skew defaults to 300s when <= 0.
func NewMerchantSign(opt MerchantSignOptions) *MerchantSign {
	skew := opt.AllowedSkewSeconds
	if skew <= 0 {
		skew = 300
	}
	return &MerchantSign{
		lookup:               opt.Lookup,
		replayGuard:          opt.ReplayGuard,
		allowedSkewSeconds:   skew,
		trustForwardedForIPs: opt.TrustForwardedForIPs,
		fail:                 opt.Fail,
		codes:                opt.Codes,
	}
}

func (m *MerchantSign) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := ReadMergedParams(r)
		if err != nil {
			m.fail(w, m.codes.InvalidParams, "invalid params")
			return
		}
		appID := params["app_id"]
		if appID == "" {
			m.fail(w, m.codes.AppIDRequired, "app_id required")
			return
		}
		sign := params["sign"]
		if sign == "" {
			m.fail(w, m.codes.SignRequired, "sign required")
			return
		}
		tsRaw := strings.TrimSpace(params["timestamp"])
		if tsRaw == "" {
			m.fail(w, m.codes.TimestampRequired, "timestamp required")
			return
		}
		ts, err := strconv.ParseInt(tsRaw, 10, 64)
		if err != nil {
			m.fail(w, m.codes.InvalidTimestamp, "invalid timestamp")
			return
		}
		now := time.Now().Unix()
		if ts < now-m.allowedSkewSeconds || ts > now+m.allowedSkewSeconds {
			m.fail(w, m.codes.InvalidTimestamp, "timestamp out of allowed window")
			return
		}
		nonce := strings.TrimSpace(params["nonce"])
		if nonce == "" {
			m.fail(w, m.codes.NonceRequired, "nonce required")
			return
		}

		acct, err := m.lookup.LookupByAppID(r.Context(), appID)
		if err != nil {
			if errors.Is(err, ErrMerchantSignNotFound) {
				m.fail(w, m.codes.MerchantNotFound, "merchant not found")
				return
			}
			m.fail(w, m.codes.Internal, "merchant lookup failed")
			return
		}
		if !acct.Active {
			m.fail(w, m.codes.MerchantDisabled, "merchant disabled")
			return
		}
		clientHost := ClientHost(r, m.trustForwardedForIPs)
		if !IPAllowed(clientHost, acct.IPWhitelist) {
			m.fail(w, m.codes.IPNotAllowed, "ip not allowed")
			return
		}

		expect := signmd5.SignSortedKV(params, acct.AppSecret)
		if !strings.EqualFold(expect, sign) {
			m.fail(w, m.codes.InvalidSign, "invalid sign")
			return
		}
		merchantID := strings.TrimSpace(acct.MerchantID)
		if merchantID == "" {
			m.fail(w, m.codes.MerchantNotFound, "merchant not found")
			return
		}
		params["merchant_id"] = merchantID
		if m.replayGuard != nil {
			ok, err := m.replayGuard.MarkSeen(r.Context(), merchantID, nonce, ts)
			if err != nil {
				m.fail(w, m.codes.Unavailable, "replay guard unavailable")
				return
			}
			if !ok {
				m.fail(w, m.codes.ReplayRequest, "replay request rejected")
				return
			}
		}

		next(w, r)
	}
}
