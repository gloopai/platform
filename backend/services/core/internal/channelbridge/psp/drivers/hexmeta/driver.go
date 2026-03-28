// Package hexmeta: India Hexmeta-style PSP (docs/in/README.md).
package hexmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
)

const apiPrefix = "/exposed/v1"

// DriverKey is channels.payin_type for this implementation.
const DriverKey = "hexmeta_in"

type cfg struct {
	GatewayURL string
	AppID      string
	Secret     string
}

type driver struct {
	contracts.BaseChannelDriver
	cfg    *cfg
	client *http.Client
}

// NewDriver loads channel row via DB + KV snapshot (same as Resolver), merges JSON in hexmeta, then [parseConfig].
func NewDriver(channelID int64, ch *store.ChannelsStore, snap *kvcache.ChannelSnapshot) (contracts.ChannelDriver, error) {
	if channelID <= 0 {
		return nil, fmt.Errorf("hexmeta: invalid channel_id")
	}
	merged, err := CanonicalBindJSONFromKV(ch, snap, channelID)
	if err != nil {
		return nil, err
	}
	c, err := parseConfig(merged)
	if err != nil {
		return nil, err
	}
	return &driver{
		BaseChannelDriver: contracts.NewBaseChannelDriver(channelID, DriverKey),
		cfg:               c,
		client:            &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (d *driver) DriverKey() string { return DriverKey }
func (d *driver) ChannelID() int64  { return d.BaseChannelDriver.ChannelID }

func (d *driver) CreatePayment(ctx context.Context, req *contracts.CreatePaymentReq) (*contracts.CreatePaymentResp, error) {
	if req == nil {
		return nil, fmt.Errorf("hexmeta: nil CreatePaymentReq")
	}
	body := map[string]string{
		"orderNo":   req.MerchantOrderNo,
		"amount":    strconv.FormatInt(req.AmountMinor, 10),
		"name":      fillPayerName(req.PayerName),
		"phone":     fillPayerPhone(req.PayerPhone),
		"email":     fillPayerEmail(req.PayerEmail),
		"notifyUrl": strings.TrimSpace(req.NotifyURL),
	}
	if ip := strings.TrimSpace(req.UserIP); ip != "" {
		body["userIP"] = ip
	}
	raw, err := d.postJSON(ctx, "/order/payment", body)
	if err != nil {
		return nil, err
	}
	var data struct {
		SysOrderNo string `json:"sysOrderNo"`
		PayURL     string `json:"payUrl"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("hexmeta: decode create payment data: %w", err)
	}
	return &contracts.CreatePaymentResp{
		ChannelOrderNo: strings.TrimSpace(data.SysOrderNo),
		PayURL:         strings.TrimSpace(data.PayURL),
	}, nil
}

func (d *driver) QueryPayment(ctx context.Context, req *contracts.QueryPaymentReq) (*contracts.QueryPaymentResp, error) {
	if req == nil {
		return nil, fmt.Errorf("hexmeta: nil QueryPaymentReq")
	}
	body := map[string]string{"orderNo": req.MerchantOrderNo}
	raw, err := d.postJSON(ctx, "/query/payment", body)
	if err != nil {
		return nil, err
	}
	var data struct {
		AppID       string `json:"appId"`
		OrderNo     string `json:"orderNo"`
		SysOrderNo  string `json:"sysOrderNo"`
		Amount      string `json:"amount"`
		Status      string `json:"status"`
		ReferenceNo string `json:"referenceNo"`
		FailReason  string `json:"failReason"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("hexmeta: decode query payment: %w", err)
	}
	amt, _ := strconv.ParseInt(strings.TrimSpace(data.Amount), 10, 64)
	st := strings.TrimSpace(data.Status)
	return &contracts.QueryPaymentResp{
		AppID:           strings.TrimSpace(data.AppID),
		MerchantOrderNo: strings.TrimSpace(data.OrderNo),
		ChannelOrderNo:  strings.TrimSpace(data.SysOrderNo),
		AmountMinor:     amt,
		Status:          parsePayinStatus(st),
		ReferenceNo:     strings.TrimSpace(data.ReferenceNo),
		FailReason:      strings.TrimSpace(data.FailReason),
		RawStatus:       st,
	}, nil
}

func (d *driver) Makeup(ctx context.Context, req *contracts.MakeupReq) error {
	if req == nil {
		return fmt.Errorf("hexmeta: nil MakeupReq")
	}
	body := map[string]string{
		"orderNo":     req.MerchantOrderNo,
		"referenceNo": strings.TrimSpace(req.ReferenceNo),
	}
	_, err := d.postJSON(ctx, "/makeup", body)
	return err
}

func (d *driver) VerifyPayinNotify(ctx context.Context, r *http.Request) (*contracts.PayinNotifyParsed, error) {
	_ = ctx
	raw, err := readNotifyBody(r)
	if err != nil {
		return nil, err
	}
	m, err := verifyNotifyBody(raw, d.cfg.Secret)
	if err != nil {
		return nil, err
	}
	amt, _ := strconv.ParseInt(strings.TrimSpace(m["amount"]), 10, 64)
	st := strings.TrimSpace(m["status"])
	return &contracts.PayinNotifyParsed{
		MerchantOrderNo: strings.TrimSpace(m["orderNo"]),
		ChannelOrderNo:  strings.TrimSpace(m["sysOrderNo"]),
		PaidAmountMinor: amt,
		Status:          parsePayinStatus(st),
		RawStatus:       st,
	}, nil
}

func (d *driver) CreatePayout(ctx context.Context, req *contracts.CreatePayoutReq) (*contracts.CreatePayoutResp, error) {
	if req == nil {
		return nil, fmt.Errorf("hexmeta: nil CreatePayoutReq")
	}
	bankName := strings.TrimSpace(req.BankName)
	if bankName == "" {
		bankName = "IndiaBank"
	}
	body := map[string]string{
		"orderNo":   req.MerchantOrderNo,
		"wayCode":   wayCodeStr(req.WayCode),
		"amount":    strconv.FormatInt(req.AmountMinor, 10),
		"bankName":  bankName,
		"bankCode":  strings.TrimSpace(req.BankCode),
		"accountNo": strings.TrimSpace(req.AccountNo),
		"name":      fillPayerName(req.HolderName),
		"phone":     fillPayerPhone(req.Phone),
		"email":     fillPayerEmail(req.Email),
		"notifyUrl": strings.TrimSpace(req.NotifyURL),
	}
	raw, err := d.postJSON(ctx, "/order/payout", body)
	if err != nil {
		return nil, err
	}
	var data struct {
		SysOrderNo string `json:"sysOrderNo"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("hexmeta: decode create payout: %w", err)
	}
	return &contracts.CreatePayoutResp{ChannelOrderNo: strings.TrimSpace(data.SysOrderNo)}, nil
}

func (d *driver) QueryPayout(ctx context.Context, req *contracts.QueryPayoutReq) (*contracts.QueryPayoutResp, error) {
	_ = ctx
	_ = req
	return nil, contracts.ErrUnsupported
}

func (d *driver) VerifyPayoutNotify(ctx context.Context, r *http.Request) (*contracts.PayoutNotifyParsed, error) {
	_ = ctx
	raw, err := readNotifyBody(r)
	if err != nil {
		return nil, err
	}
	m, err := verifyNotifyBody(raw, d.cfg.Secret)
	if err != nil {
		return nil, err
	}
	amt, _ := strconv.ParseInt(strings.TrimSpace(m["amount"]), 10, 64)
	st := strings.TrimSpace(m["status"])
	return &contracts.PayoutNotifyParsed{
		MerchantOrderNo: strings.TrimSpace(m["orderNo"]),
		ChannelOrderNo:  strings.TrimSpace(m["sysOrderNo"]),
		AmountMinor:     amt,
		Status:          parsePayoutStatus(st),
		ReferenceNo:     strings.TrimSpace(m["referenceNo"]),
		RawStatus:       st,
	}, nil
}

func (d *driver) QueryBalance(ctx context.Context) (*contracts.BalanceSnapshot, error) {
	body := map[string]string{}
	raw, err := d.postJSON(ctx, "/query/balance", body)
	if err != nil {
		return nil, err
	}
	var data struct {
		AvailableBalance string `json:"availableBalance"`
		UnsettledAmount  string `json:"unsettledAmount"`
		FrozenAmount     string `json:"frozenAmount"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("hexmeta: decode balance: %w", err)
	}
	avail, _ := strconv.ParseInt(strings.TrimSpace(data.AvailableBalance), 10, 64)
	unset, _ := strconv.ParseInt(strings.TrimSpace(data.UnsettledAmount), 10, 64)
	frozen, _ := strconv.ParseInt(strings.TrimSpace(data.FrozenAmount), 10, 64)
	return &contracts.BalanceSnapshot{
		AvailableMinor: avail,
		UnsettledMinor: unset,
		FrozenMinor:    frozen,
	}, nil
}
