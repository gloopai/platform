package hexmeta

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/pay/core/internal/channelbind/channeldriver/base"
)

// DriverKey is the channels.payin_type / driver_key value for this PSP (India Hexmeta-style API).
const DriverKey = "hexmeta_in"

// Driver implements [base.ChannelDriver] for the upstream described in docs/in/README.md.
type Driver struct {
	base.BaseChannelDriver
	cfg    *Config
	client *http.Client
}

// New parses channel_config and returns a bound driver.
func New(channelID int64, bindJSON string) (*Driver, error) {
	cfg, err := parseConfig(bindJSON)
	if err != nil {
		return nil, err
	}
	return &Driver{
		BaseChannelDriver: base.NewBaseChannelDriver(channelID, DriverKey),
		cfg:               cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (d *Driver) DriverKey() string { return DriverKey }

func (d *Driver) ChannelID() int64 { return d.BaseChannelDriver.ChannelID }

func (d *Driver) CreatePayment(ctx context.Context, req *base.CreatePaymentReq) (*base.CreatePaymentResp, error) {
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
	return &base.CreatePaymentResp{
		ChannelOrderNo: strings.TrimSpace(data.SysOrderNo),
		PayURL:         strings.TrimSpace(data.PayURL),
	}, nil
}

func (d *Driver) QueryPayment(ctx context.Context, req *base.QueryPaymentReq) (*base.QueryPaymentResp, error) {
	if req == nil {
		return nil, fmt.Errorf("hexmeta: nil QueryPaymentReq")
	}
	body := map[string]string{
		"orderNo": req.MerchantOrderNo,
	}
	raw, err := d.postJSON(ctx, "/query/payment", body)
	if err != nil {
		return nil, err
	}
	var data struct {
		AppID        string `json:"appId"`
		OrderNo      string `json:"orderNo"`
		SysOrderNo   string `json:"sysOrderNo"`
		Amount       string `json:"amount"`
		Status       string `json:"status"`
		ReferenceNo  string `json:"referenceNo"`
		FailReason   string `json:"failReason"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("hexmeta: decode query payment: %w", err)
	}
	amt, _ := strconv.ParseInt(strings.TrimSpace(data.Amount), 10, 64)
	st := strings.TrimSpace(data.Status)
	return &base.QueryPaymentResp{
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

func (d *Driver) Makeup(ctx context.Context, req *base.MakeupReq) error {
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

func (d *Driver) VerifyPayinNotify(ctx context.Context, r *http.Request) (*base.PayinNotifyParsed, error) {
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
	return &base.PayinNotifyParsed{
		MerchantOrderNo: strings.TrimSpace(m["orderNo"]),
		ChannelOrderNo:  strings.TrimSpace(m["sysOrderNo"]),
		PaidAmountMinor: amt,
		Status:          parsePayinStatus(st),
		RawStatus:       st,
	}, nil
}

func (d *Driver) CreatePayout(ctx context.Context, req *base.CreatePayoutReq) (*base.CreatePayoutResp, error) {
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
	return &base.CreatePayoutResp{ChannelOrderNo: strings.TrimSpace(data.SysOrderNo)}, nil
}

func (d *Driver) QueryPayout(ctx context.Context, req *base.QueryPayoutReq) (*base.QueryPayoutResp, error) {
	_ = ctx
	_ = req
	return nil, base.ErrUnsupported
}

func (d *Driver) VerifyPayoutNotify(ctx context.Context, r *http.Request) (*base.PayoutNotifyParsed, error) {
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
	return &base.PayoutNotifyParsed{
		MerchantOrderNo: strings.TrimSpace(m["orderNo"]),
		ChannelOrderNo:  strings.TrimSpace(m["sysOrderNo"]),
		AmountMinor:     amt,
		Status:          parsePayoutStatus(st),
		ReferenceNo:     strings.TrimSpace(m["referenceNo"]),
		RawStatus:       st,
	}, nil
}

func (d *Driver) QueryBalance(ctx context.Context) (*base.BalanceSnapshot, error) {
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
	return &base.BalanceSnapshot{
		AvailableMinor: avail,
		UnsettledMinor: unset,
		FrozenMinor:    frozen,
	}, nil
}

func parsePayinStatus(s string) base.PayinOrderStatus {
	switch s {
	case "1":
		return base.PayinStatusProcessing
	case "2":
		return base.PayinStatusSuccess
	case "3":
		return base.PayinStatusFailed
	default:
		return base.PayinStatusUnknown
	}
}

func parsePayoutStatus(s string) base.PayoutOrderStatus {
	switch s {
	case "1":
		return base.PayoutStatusProcessing
	case "2":
		return base.PayoutStatusSuccess
	case "3":
		return base.PayoutStatusFailed
	default:
		return base.PayoutStatusUnknown
	}
}

func wayCodeStr(w base.PayoutWayCode) string {
	switch w {
	case base.PayoutWayUPI:
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
