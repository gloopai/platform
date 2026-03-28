package mockpsp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gloopai/pay/channeldriver"
)

// DefaultDriverKey is the registry key unless overridden in New.
const DefaultDriverKey = "mock_psp"

// Driver is an in-memory mock PSP implementing PayinChannel, PayoutChannel, and BalanceChannel.
type Driver struct {
	key   string
	store *Store
	seq   int64
}

// New returns a Driver with an isolated Store. Empty key uses DefaultDriverKey.
func New(key string) *Driver {
	if key == "" {
		key = DefaultDriverKey
	}
	return &Driver{key: key, store: NewStore()}
}

// Store exposes the backing store for test assertions.
func (d *Driver) Store() *Store { return d.store }

func (d *Driver) Key() string { return d.key }

// --- Payin ---

func (d *Driver) CreatePayment(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.CreatePaymentReq) (*channeldriver.CreatePaymentResp, error) {
	_ = ctx
	if cfg == nil || req == nil {
		return nil, errors.New("mockpsp: nil cfg or req")
	}
	if req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp: merchant order no required")
	}
	sys := fmt.Sprintf("MOCK%d", atomic.AddInt64(&d.seq, 1))
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	if _, dup := d.store.payin[req.MerchantOrderNo]; dup {
		return nil, errors.New("mockpsp: duplicate merchant order")
	}
	d.store.payin[req.MerchantOrderNo] = &payinRec{
		sysOrderNo:      sys,
		merchantOrderNo: req.MerchantOrderNo,
		amountMinor:     req.AmountMinor,
		status:          channeldriver.PayinStatusProcessing,
	}
	payURL := fmt.Sprintf("https://mock.psp.test/pay?merchantOrderNo=%s&sysOrderNo=%s", req.MerchantOrderNo, sys)
	return &channeldriver.CreatePaymentResp{ChannelOrderNo: sys, PayURL: payURL}, nil
}

func (d *Driver) QueryPayment(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.QueryPaymentReq) (*channeldriver.QueryPaymentResp, error) {
	_ = ctx
	appID := ""
	if cfg != nil {
		appID = cfg.AppID
	}
	if req == nil || req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp: merchant order no required")
	}
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	r, ok := d.store.payin[req.MerchantOrderNo]
	if !ok {
		return nil, errors.New("mockpsp: order not found")
	}
	return &channeldriver.QueryPaymentResp{
		AppID:           appID,
		MerchantOrderNo: r.merchantOrderNo,
		ChannelOrderNo: r.sysOrderNo,
		AmountMinor:     r.amountMinor,
		Status:          r.status,
		ReferenceNo:     r.referenceNo,
		FailReason:      r.failReason,
		RawStatus:       payinStatusString(r.status),
	}, nil
}

func payinStatusString(s channeldriver.PayinOrderStatus) string {
	switch s {
	case channeldriver.PayinStatusProcessing:
		return "1"
	case channeldriver.PayinStatusSuccess:
		return "2"
	case channeldriver.PayinStatusFailed:
		return "3"
	default:
		return "0"
	}
}

func (d *Driver) Makeup(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.MakeupReq) error {
	_ = ctx
	_ = cfg
	if req == nil || req.MerchantOrderNo == "" || req.ReferenceNo == "" {
		return errors.New("mockpsp: makeup orderNo and referenceNo required")
	}
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	r, ok := d.store.payin[req.MerchantOrderNo]
	if !ok {
		return errors.New("mockpsp: order not found")
	}
	r.referenceNo = req.ReferenceNo
	r.status = channeldriver.PayinStatusSuccess
	return nil
}

// SetPayinFailed sets mock payin to failed (for query simulation / tests).
func (d *Driver) SetPayinFailed(merchantOrderNo, failReason string) error {
	if merchantOrderNo == "" {
		return errors.New("mockpsp: merchant order no required")
	}
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	r, ok := d.store.payin[merchantOrderNo]
	if !ok {
		return errors.New("mockpsp: order not found")
	}
	r.status = channeldriver.PayinStatusFailed
	r.failReason = failReason
	return nil
}

// SetPayoutStatus sets mock payout status (for query simulation / tests).
func (d *Driver) SetPayoutStatus(merchantOrderNo string, status channeldriver.PayoutOrderStatus, referenceNo string) error {
	if merchantOrderNo == "" {
		return errors.New("mockpsp: merchant order no required")
	}
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	r, ok := d.store.payout[merchantOrderNo]
	if !ok {
		return errors.New("mockpsp: payout order not found")
	}
	r.status = status
	r.referenceNo = referenceNo
	return nil
}

type payinNotifyJSON struct {
	Timestamp string `json:"timestamp"`
	Sign      string `json:"sign"`
	OrderNo   string `json:"orderNo"`
	SysOrderNo string `json:"sysOrderNo"`
	Status    string `json:"status"`
	Amount    string `json:"amount"`
}

func (d *Driver) VerifyPayinNotify(ctx context.Context, cfg *channeldriver.ChannelConfig, r *http.Request) (*channeldriver.PayinNotifyParsed, error) {
	_ = ctx
	if cfg == nil || r == nil {
		return nil, channeldriver.ErrVerifyNotify
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var j payinNotifyJSON
	if err := json.Unmarshal(body, &j); err != nil {
		return nil, fmt.Errorf("%w: %v", channeldriver.ErrVerifyNotify, err)
	}
	fields := map[string]string{
		"timestamp":  j.Timestamp,
		"orderNo":    j.OrderNo,
		"sysOrderNo": j.SysOrderNo,
		"status":     j.Status,
		"amount":     j.Amount,
	}
	want := SignHMAC(cfg.SignSecret, fields)
	if !hmacEqual(want, j.Sign) {
		return nil, channeldriver.ErrVerifyNotify
	}
	amt, err := strconv.ParseInt(j.Amount, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: amount: %v", channeldriver.ErrVerifyNotify, err)
	}
	st := parsePayinStatus(j.Status)
	return &channeldriver.PayinNotifyParsed{
		MerchantOrderNo: j.OrderNo,
		ChannelOrderNo: j.SysOrderNo,
		PaidAmountMinor: amt,
		Status:          st,
		RawStatus:       j.Status,
	}, nil
}

func parsePayinStatus(s string) channeldriver.PayinOrderStatus {
	switch s {
	case "1":
		return channeldriver.PayinStatusProcessing
	case "2":
		return channeldriver.PayinStatusSuccess
	case "3":
		return channeldriver.PayinStatusFailed
	default:
		return channeldriver.PayinStatusUnknown
	}
}

func (d *Driver) PayinNotifyResponse(success bool) []byte {
	if success {
		return []byte("SUCCESS")
	}
	return []byte("FAIL")
}

// --- Payout ---

func (d *Driver) CreatePayout(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.CreatePayoutReq) (*channeldriver.CreatePayoutResp, error) {
	_ = ctx
	if cfg == nil || req == nil {
		return nil, errors.New("mockpsp: nil cfg or req")
	}
	if req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp: merchant order no required")
	}
	sys := fmt.Sprintf("MOCKPO%d", atomic.AddInt64(&d.seq, 1))
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	if _, dup := d.store.payout[req.MerchantOrderNo]; dup {
		return nil, errors.New("mockpsp: duplicate payout merchant order")
	}
	d.store.payout[req.MerchantOrderNo] = &payoutRec{
		sysOrderNo:      sys,
		merchantOrderNo: req.MerchantOrderNo,
		amountMinor:     req.AmountMinor,
		status:          channeldriver.PayoutStatusProcessing,
	}
	return &channeldriver.CreatePayoutResp{ChannelOrderNo: sys}, nil
}

func (d *Driver) QueryPayout(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.QueryPayoutReq) (*channeldriver.QueryPayoutResp, error) {
	_ = ctx
	_ = cfg
	if req == nil || req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp: merchant order no required")
	}
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	r, ok := d.store.payout[req.MerchantOrderNo]
	if !ok {
		return nil, errors.New("mockpsp: payout order not found")
	}
	return &channeldriver.QueryPayoutResp{
		MerchantOrderNo: r.merchantOrderNo,
		ChannelOrderNo: r.sysOrderNo,
		AmountMinor:     r.amountMinor,
		Status:          r.status,
		ReferenceNo:     r.referenceNo,
		RawStatus:       payoutStatusString(r.status),
	}, nil
}

func payoutStatusString(s channeldriver.PayoutOrderStatus) string {
	switch s {
	case channeldriver.PayoutStatusProcessing:
		return "1"
	case channeldriver.PayoutStatusSuccess:
		return "2"
	case channeldriver.PayoutStatusFailed:
		return "3"
	default:
		return "0"
	}
}

type payoutNotifyJSON struct {
	OrderNo     string `json:"orderNo"`
	SysOrderNo  string `json:"sysOrderNo"`
	Status      string `json:"status"`
	Amount      string `json:"amount"`
	ReferenceNo string `json:"referenceNo"`
	Timestamp   string `json:"timestamp"`
	Sign        string `json:"sign"`
}

func parsePayoutStatus(s string) channeldriver.PayoutOrderStatus {
	switch s {
	case "1":
		return channeldriver.PayoutStatusProcessing
	case "2":
		return channeldriver.PayoutStatusSuccess
	case "3":
		return channeldriver.PayoutStatusFailed
	default:
		return channeldriver.PayoutStatusUnknown
	}

}

func (d *Driver) VerifyPayoutNotify(ctx context.Context, cfg *channeldriver.ChannelConfig, r *http.Request) (*channeldriver.PayoutNotifyParsed, error) {
	_ = ctx
	if cfg == nil || r == nil {
		return nil, channeldriver.ErrVerifyNotify
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var j payoutNotifyJSON
	if err := json.Unmarshal(body, &j); err != nil {
		return nil, fmt.Errorf("%w: %v", channeldriver.ErrVerifyNotify, err)
	}
	fields := map[string]string{
		"orderNo":     j.OrderNo,
		"sysOrderNo":  j.SysOrderNo,
		"status":      j.Status,
		"amount":      j.Amount,
		"referenceNo": j.ReferenceNo,
		"timestamp":   j.Timestamp,
	}
	want := SignHMAC(cfg.SignSecret, fields)
	if !hmacEqual(want, j.Sign) {
		return nil, channeldriver.ErrVerifyNotify
	}
	amt, err := strconv.ParseInt(j.Amount, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: amount: %v", channeldriver.ErrVerifyNotify, err)
	}
	st := parsePayoutStatus(j.Status)
	return &channeldriver.PayoutNotifyParsed{
		MerchantOrderNo: j.OrderNo,
		ChannelOrderNo: j.SysOrderNo,
		AmountMinor:     amt,
		Status:          st,
		ReferenceNo:     j.ReferenceNo,
		RawStatus:       j.Status,
	}, nil
}

func (d *Driver) PayoutNotifyResponse(success bool) []byte {
	return d.PayinNotifyResponse(success)
}

func (d *Driver) QueryBalance(ctx context.Context, cfg *channeldriver.ChannelConfig) (*channeldriver.BalanceSnapshot, error) {
	_ = ctx
	_ = cfg
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	return &channeldriver.BalanceSnapshot{
		AvailableMinor: d.store.availMinor,
		UnsettledMinor: d.store.unsettledMinor,
		FrozenMinor:    d.store.frozenMinor,
	}, nil
}

func hmacEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// BuildPayinNotifyBody returns JSON body bytes and Content-Type application/json for posting to your gateway test URL.
func BuildPayinNotifyBody(cfg *channeldriver.ChannelConfig, merchantOrderNo, sysOrderNo string, status channeldriver.PayinOrderStatus, amountMinor int64) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("mockpsp: nil cfg")
	}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	st := payinStatusString(status)
	amt := strconv.FormatInt(amountMinor, 10)
	fields := map[string]string{
		"timestamp":  ts,
		"orderNo":    merchantOrderNo,
		"sysOrderNo": sysOrderNo,
		"status":     st,
		"amount":     amt,
	}
	sig := SignHMAC(cfg.SignSecret, fields)
	j := payinNotifyJSON{
		Timestamp:  ts,
		Sign:       sig,
		OrderNo:    merchantOrderNo,
		SysOrderNo: sysOrderNo,
		Status:     st,
		Amount:     amt,
	}
	return json.Marshal(j)
}

// BuildPayoutNotifyBody builds payout async notify JSON (matches mock VerifyPayoutNotify field set).
func BuildPayoutNotifyBody(cfg *channeldriver.ChannelConfig, merchantOrderNo, sysOrderNo string, status channeldriver.PayoutOrderStatus, amountMinor int64, utr string) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("mockpsp: nil cfg")
	}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	st := payoutStatusString(status)
	amt := strconv.FormatInt(amountMinor, 10)
	fields := map[string]string{
		"orderNo":     merchantOrderNo,
		"sysOrderNo":  sysOrderNo,
		"status":      st,
		"amount":      amt,
		"referenceNo": utr,
		"timestamp":   ts,
	}
	sig := SignHMAC(cfg.SignSecret, fields)
	j := payoutNotifyJSON{
		OrderNo:     merchantOrderNo,
		SysOrderNo:  sysOrderNo,
		Status:      st,
		Amount:      amt,
		ReferenceNo: utr,
		Timestamp:   ts,
		Sign:        sig,
	}
	return json.Marshal(j)
}
