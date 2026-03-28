package mockpsp2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gloopai/pay/channeldriver"
)

// DefaultDriverKey 与 channels.payin_type / 注册表一致。
const DefaultDriverKey = "mock_psp_alt"

// Driver 为内存 mock 上游：字段 snake_case + MD5 签名，与 mock_psp（camelCase + HMAC-SHA256）区分。
type Driver struct {
	key   string
	store *Store
	seq   int64
}

// New 返回独立 Store 的 Driver；key 为空则用 DefaultDriverKey。
func New(key string) *Driver {
	if key == "" {
		key = DefaultDriverKey
	}
	return &Driver{key: key, store: NewStore()}
}

// Store 暴露内存态供测试断言。
func (d *Driver) Store() *Store { return d.store }

func (d *Driver) Key() string { return d.key }

// --- Payin ---

func (d *Driver) CreatePayment(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.CreatePaymentReq) (*channeldriver.CreatePaymentResp, error) {
	_ = ctx
	if cfg == nil || req == nil {
		return nil, errors.New("mockpsp2: nil cfg or req")
	}
	if req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp2: merchant order no required")
	}
	sys := fmt.Sprintf("ALT%d", atomic.AddInt64(&d.seq, 1))
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	if _, dup := d.store.payin[req.MerchantOrderNo]; dup {
		return nil, errors.New("mockpsp2: duplicate merchant order")
	}
	d.store.payin[req.MerchantOrderNo] = &payinRec{
		sysOrderNo:      sys,
		merchantOrderNo: req.MerchantOrderNo,
		amountMinor:     req.AmountMinor,
		status:          channeldriver.PayinStatusProcessing,
	}
	payURL := fmt.Sprintf("https://alt-mock.psp.test/cashier?merchant_ref=%s&txn_id=%s", req.MerchantOrderNo, sys)
	return &channeldriver.CreatePaymentResp{UpstreamOrderNo: sys, PayURL: payURL}, nil
}

func (d *Driver) QueryPayment(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.QueryPaymentReq) (*channeldriver.QueryPaymentResp, error) {
	_ = ctx
	appID := ""
	if cfg != nil {
		appID = cfg.AppID
	}
	if req == nil || req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp2: merchant order no required")
	}
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	r, ok := d.store.payin[req.MerchantOrderNo]
	if !ok {
		return nil, errors.New("mockpsp2: order not found")
	}
	return &channeldriver.QueryPaymentResp{
		AppID:           appID,
		MerchantOrderNo: r.merchantOrderNo,
		UpstreamOrderNo: r.sysOrderNo,
		AmountMinor:     r.amountMinor,
		Status:          r.status,
		ReferenceNo:     r.referenceNo,
		FailReason:      r.failReason,
		RawStatus:       payinStatusWord(r.status),
	}, nil
}

func payinStatusWord(s channeldriver.PayinOrderStatus) string {
	switch s {
	case channeldriver.PayinStatusProcessing:
		return "PENDING"
	case channeldriver.PayinStatusSuccess:
		return "SUCCESS"
	case channeldriver.PayinStatusFailed:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

func (d *Driver) Makeup(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.MakeupReq) error {
	_ = ctx
	_ = cfg
	if req == nil || req.MerchantOrderNo == "" || req.ReferenceNo == "" {
		return errors.New("mockpsp2: makeup orderNo and referenceNo required")
	}
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	r, ok := d.store.payin[req.MerchantOrderNo]
	if !ok {
		return errors.New("mockpsp2: order not found")
	}
	r.referenceNo = req.ReferenceNo
	r.status = channeldriver.PayinStatusSuccess
	return nil
}

// payinNotifyJSON 上游异步代收（与 mock_psp 字段名不同）。
type payinNotifyJSON struct {
	MerchantRef string `json:"merchant_ref"`
	TxnID       string `json:"txn_id"`
	State       string `json:"state"`
	Amount      string `json:"amount"`
	EventTime   string `json:"event_time"`
	Signature   string `json:"signature"`
}

func parsePayinState(s string) channeldriver.PayinOrderStatus {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "PENDING":
		return channeldriver.PayinStatusProcessing
	case "SUCCESS":
		return channeldriver.PayinStatusSuccess
	case "FAIL", "FAILED":
		return channeldriver.PayinStatusFailed
	default:
		return channeldriver.PayinStatusUnknown
	}
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
	signParams := map[string]string{
		"amount":       strings.TrimSpace(j.Amount),
		"event_time":   strings.TrimSpace(j.EventTime),
		"merchant_ref": strings.TrimSpace(j.MerchantRef),
		"state":        strings.TrimSpace(j.State),
		"txn_id":       strings.TrimSpace(j.TxnID),
	}
	want := SignMd5SortedKV(signParams, cfg.SignSecret)
	if !strings.EqualFold(want, strings.TrimSpace(j.Signature)) {
		return nil, channeldriver.ErrVerifyNotify
	}
	amt, err := strconv.ParseInt(strings.TrimSpace(j.Amount), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: amount: %v", channeldriver.ErrVerifyNotify, err)
	}
	st := parsePayinState(j.State)
	return &channeldriver.PayinNotifyParsed{
		MerchantOrderNo: strings.TrimSpace(j.MerchantRef),
		UpstreamOrderNo: strings.TrimSpace(j.TxnID),
		PaidAmountMinor: amt,
		Status:          st,
		RawStatus:       strings.TrimSpace(j.State),
	}, nil
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
		return nil, errors.New("mockpsp2: nil cfg or req")
	}
	if req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp2: merchant order no required")
	}
	sys := fmt.Sprintf("ALTPO%d", atomic.AddInt64(&d.seq, 1))
	d.store.mu.Lock()
	defer d.store.mu.Unlock()
	if _, dup := d.store.payout[req.MerchantOrderNo]; dup {
		return nil, errors.New("mockpsp2: duplicate payout merchant order")
	}
	d.store.payout[req.MerchantOrderNo] = &payoutRec{
		sysOrderNo:      sys,
		merchantOrderNo: req.MerchantOrderNo,
		amountMinor:     req.AmountMinor,
		status:          channeldriver.PayoutStatusProcessing,
	}
	return &channeldriver.CreatePayoutResp{UpstreamOrderNo: sys}, nil
}

func (d *Driver) QueryPayout(ctx context.Context, cfg *channeldriver.ChannelConfig, req *channeldriver.QueryPayoutReq) (*channeldriver.QueryPayoutResp, error) {
	_ = ctx
	_ = cfg
	if req == nil || req.MerchantOrderNo == "" {
		return nil, errors.New("mockpsp2: merchant order no required")
	}
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	r, ok := d.store.payout[req.MerchantOrderNo]
	if !ok {
		return nil, errors.New("mockpsp2: payout order not found")
	}
	return &channeldriver.QueryPayoutResp{
		MerchantOrderNo: r.merchantOrderNo,
		UpstreamOrderNo: r.sysOrderNo,
		AmountMinor:     r.amountMinor,
		Status:          r.status,
		ReferenceNo:     r.referenceNo,
		RawStatus:       payoutStatusWord(r.status),
	}, nil
}

func payoutStatusWord(s channeldriver.PayoutOrderStatus) string {
	switch s {
	case channeldriver.PayoutStatusProcessing:
		return "PROCESSING"
	case channeldriver.PayoutStatusSuccess:
		return "SUCCESS"
	case channeldriver.PayoutStatusFailed:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

type payoutNotifyJSON struct {
	MerchantRef    string `json:"merchant_ref"`
	TxnID          string `json:"txn_id"`
	PayoutState    string `json:"payout_state"`
	Amount         string `json:"amount"`
	BankReference  string `json:"bank_reference"`
	EventTime      string `json:"event_time"`
	Signature      string `json:"signature"`
}

func parsePayoutState(s string) channeldriver.PayoutOrderStatus {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "PROCESSING":
		return channeldriver.PayoutStatusProcessing
	case "SUCCESS":
		return channeldriver.PayoutStatusSuccess
	case "FAIL", "FAILED":
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
	signParams := map[string]string{
		"amount":          strings.TrimSpace(j.Amount),
		"bank_reference":  strings.TrimSpace(j.BankReference),
		"event_time":      strings.TrimSpace(j.EventTime),
		"merchant_ref":    strings.TrimSpace(j.MerchantRef),
		"payout_state":    strings.TrimSpace(j.PayoutState),
		"txn_id":          strings.TrimSpace(j.TxnID),
	}
	want := SignMd5SortedKV(signParams, cfg.SignSecret)
	if !strings.EqualFold(want, strings.TrimSpace(j.Signature)) {
		return nil, channeldriver.ErrVerifyNotify
	}
	amt, err := strconv.ParseInt(strings.TrimSpace(j.Amount), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: amount: %v", channeldriver.ErrVerifyNotify, err)
	}
	st := parsePayoutState(j.PayoutState)
	return &channeldriver.PayoutNotifyParsed{
		MerchantOrderNo: strings.TrimSpace(j.MerchantRef),
		UpstreamOrderNo: strings.TrimSpace(j.TxnID),
		AmountMinor:     amt,
		Status:          st,
		ReferenceNo:     strings.TrimSpace(j.BankReference),
		RawStatus:       strings.TrimSpace(j.PayoutState),
	}, nil
}

func (d *Driver) PayoutNotifyResponse(success bool) []byte {
	if success {
		return []byte("SUCCESS")
	}
	return []byte("FAIL")
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

// BuildPayinNotifyBody 构造代收回调 JSON（联调/脚本用）。
func BuildPayinNotifyBody(cfg *channeldriver.ChannelConfig, merchantRef, txnID string, st channeldriver.PayinOrderStatus, amountMinor int64) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("mockpsp2: nil cfg")
	}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	state := payinStatusWord(st)
	amt := strconv.FormatInt(amountMinor, 10)
	signParams := map[string]string{
		"amount":       amt,
		"event_time":   ts,
		"merchant_ref": merchantRef,
		"state":        state,
		"txn_id":       txnID,
	}
	sig := SignMd5SortedKV(signParams, cfg.SignSecret)
	j := payinNotifyJSON{
		MerchantRef: merchantRef,
		TxnID:       txnID,
		State:       state,
		Amount:      amt,
		EventTime:   ts,
		Signature:   sig,
	}
	return json.Marshal(j)
}

// BuildPayoutNotifyBody 构造代付回调 JSON。
func BuildPayoutNotifyBody(cfg *channeldriver.ChannelConfig, merchantRef, txnID string, st channeldriver.PayoutOrderStatus, amountMinor int64, bankRef string) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("mockpsp2: nil cfg")
	}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	ps := payoutStatusWord(st)
	amt := strconv.FormatInt(amountMinor, 10)
	signParams := map[string]string{
		"amount":         amt,
		"bank_reference": bankRef,
		"event_time":     ts,
		"merchant_ref":   merchantRef,
		"payout_state":   ps,
		"txn_id":         txnID,
	}
	sig := SignMd5SortedKV(signParams, cfg.SignSecret)
	j := payoutNotifyJSON{
		MerchantRef:   merchantRef,
		TxnID:         txnID,
		PayoutState:   ps,
		Amount:        amt,
		BankReference: bankRef,
		EventTime:     ts,
		Signature:     sig,
	}
	return json.Marshal(j)
}
