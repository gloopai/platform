package hexmeta

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gloopai/pay/common/model"
	"github.com/gloopai/pay/core/internal/channelbind/psp/contracts"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
)

// CanonicalBindJSONFromKV loads the channel row from DB and applies KV-picked channel_config (same as Resolver + kvcache.PickChannelConfig).
func CanonicalBindJSONFromKV(ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot, channelID int64) (string, error) {
	if ch == nil {
		return "", fmt.Errorf("hexmeta: nil channels store")
	}
	if channelID <= 0 {
		return "", fmt.Errorf("hexmeta: invalid channel_id")
	}
	ctx := context.Background()
	row, err := ch.AdminGetByID(ctx, channelID)
	if err != nil {
		return "", err
	}
	picked := kvcache.PickChannelConfig(snap, channelID, row.ChannelConfig)
	tmp := *row
	tmp.ChannelConfig = picked
	return CanonicalBindJSON(&tmp)
}

// CanonicalBindJSON merges channel_config with legacy columns and supports_* flags, then returns JSON for [parseConfig].
func CanonicalBindJSON(c *model.Channel) (string, error) {
	if c == nil {
		return "", fmt.Errorf("hexmeta: nil channel")
	}
	return channelConfigJSONForBind(c.ChannelConfig, legacyFromChannel(c), c.SupportsPayin, c.SupportsPayout)
}

type legacyRow struct {
	GatewayURL        string
	ChannelMerchantNo string
	SignSecret        string
	RSAPrivateKey     string
}

func legacyFromChannel(c *model.Channel) legacyRow {
	if c == nil {
		return legacyRow{}
	}
	return legacyRow{
		GatewayURL:        c.GatewayUrl,
		ChannelMerchantNo: c.ChannelMerchantNo,
		SignSecret:        c.SignSecret,
		RSAPrivateKey:     c.RsaPrivateKey,
	}
}

func channelConfigJSONForAPI(channelConfig string, leg legacyRow) string {
	channelConfig = strings.TrimSpace(channelConfig)
	if channelConfig != "" {
		return channelConfig
	}
	legMap := map[string]string{
		"gateway_url":         strings.TrimSpace(leg.GatewayURL),
		"channel_merchant_no": strings.TrimSpace(leg.ChannelMerchantNo),
		"sign_secret":         strings.TrimSpace(leg.SignSecret),
		"rsa_private_key":     strings.TrimSpace(leg.RSAPrivateKey),
	}
	b, err := json.Marshal(legMap)
	if err != nil {
		return ""
	}
	return string(b)
}

func channelConfigJSONForBind(channelConfig string, leg legacyRow, supportsPayin, supportsPayout bool) (string, error) {
	base := channelConfigJSONForAPI(channelConfig, leg)
	base = strings.TrimSpace(base)
	if base == "" {
		m := map[string]interface{}{
			"supports_payin":  supportsPayin,
			"supports_payout": supportsPayout,
		}
		b, err := json.Marshal(m)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(base), &m); err != nil {
		return "", err
	}
	if _, ok := m["supports_payin"]; !ok {
		m["supports_payin"] = supportsPayin
	}
	if _, ok := m["supports_payout"]; !ok {
		m["supports_payout"] = supportsPayout
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// EffectiveGatewayURL prefers gateway_url inside channel_config JSON, else the gateway_url column.
func EffectiveGatewayURL(c *model.Channel) string {
	if c == nil {
		return ""
	}
	gw := strings.TrimSpace(c.GatewayUrl)
	uc := strings.TrimSpace(c.ChannelConfig)
	if uc != "" {
		if jg := stringFromJSONObject(uc, "gateway_url"); jg != "" {
			return jg
		}
	}
	return gw
}

func stringFromJSONObject(rawJSON, key string) string {
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" || key == "" {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &m); err != nil || m == nil {
		return ""
	}
	return jsonStringField(m[key])
}

func parseConfig(raw string) (*cfg, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("hexmeta: empty channel_config")
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("hexmeta: channel_config: %w", err)
	}
	c := &cfg{
		GatewayURL: jsonStringField(m["gateway_url"]),
		AppID:      jsonStringField(m["app_id"]),
		Secret:     jsonStringField(m["sign_secret"]),
	}
	if c.AppID == "" {
		c.AppID = jsonStringField(m["channel_merchant_no"])
	}
	if c.GatewayURL == "" || c.AppID == "" || c.Secret == "" {
		return nil, fmt.Errorf("hexmeta: gateway_url, app_id (or channel_merchant_no), and sign_secret are required")
	}
	return c, nil
}

func jsonStringField(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.0f", t), "0"), ".")
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (d *driver) postJSON(ctx context.Context, path string, body map[string]string) (json.RawMessage, error) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := strings.TrimRight(d.cfg.GatewayURL, "/") + apiPrefix + path
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
		return nil, fmt.Errorf("hexmeta: HTTP %d: %s", resp.StatusCode, truncateHex(string(raw), 512))
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

func truncateHex(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func signMD5(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	b.WriteString("&key=")
	b.WriteString(secret)
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func verifyNotifyBody(body []byte, secret string) (map[string]string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var m map[string]interface{}
	if err := dec.Decode(&m); err != nil || m == nil {
		return nil, contracts.ErrVerifyNotify
	}
	sigRaw, ok := m["sign"]
	if !ok {
		return nil, contracts.ErrVerifyNotify
	}
	sigStr := strings.TrimSpace(valueForSign(sigRaw))
	if sigStr == "" {
		return nil, contracts.ErrVerifyNotify
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
		return nil, contracts.ErrVerifyNotify
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
		return nil, contracts.ErrVerifyNotify
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, contracts.ErrVerifyNotify
	}
	return raw, nil
}

func parsePayinStatus(s string) contracts.PayinOrderStatus {
	switch s {
	case "1":
		return contracts.PayinStatusProcessing
	case "2":
		return contracts.PayinStatusSuccess
	case "3":
		return contracts.PayinStatusFailed
	default:
		return contracts.PayinStatusUnknown
	}
}

func parsePayoutStatus(s string) contracts.PayoutOrderStatus {
	switch s {
	case "1":
		return contracts.PayoutStatusProcessing
	case "2":
		return contracts.PayoutStatusSuccess
	case "3":
		return contracts.PayoutStatusFailed
	default:
		return contracts.PayoutStatusUnknown
	}
}

func wayCodeStr(w contracts.PayoutWayCode) string {
	switch w {
	case contracts.PayoutWayUPI:
		return "2"
	default:
		return "1"
	}
}

func fillPayerName(s string) string {
	s = strings.TrimSpace(s)
	if s != "" {
		return s
	}
	return randomName()
}

func fillPayerPhone(s string) string {
	s = strings.TrimSpace(s)
	if s != "" {
		return s
	}
	return randomPhoneIN()
}

func fillPayerEmail(s string) string {
	s = strings.TrimSpace(s)
	if s != "" {
		return s
	}
	return randomGmail()
}

func randomDigits(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		v, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		b[i] = digits[v.Int64()]
	}
	return string(b)
}

func randomName() string {
	return "User" + randomDigits(6)
}

func randomPhoneIN() string {
	return "9" + randomDigits(9)
}

func randomGmail() string {
	return randomDigits(9) + "@gmail.com"
}
