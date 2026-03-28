package notice

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nsqio/go-nsq"
	"gorm.io/gorm"
)

// Processor consumes NSQ messages and delivers order-paid webhooks to merchants.
// It implements:
// - idempotent-ish delivery via bounded retries (retry schedule is time-based)
// - observability via writing merchant_notify_logs for every attempt
type Processor struct {
	db               *gorm.DB
	httpClient       *http.Client
	delays           []time.Duration
	maxRespBodyBytes int64
}

func NewProcessor(db *gorm.DB, httpClient *http.Client, delays []time.Duration) *Processor {
	if len(delays) == 0 {
		delays = []time.Duration{15 * time.Second, 1 * time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour}
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Processor{
		db:               db,
		httpClient:       httpClient,
		delays:           delays,
		maxRespBodyBytes: 8 << 10,
	}
}

func (p *Processor) HandleNSQMessage(msg *nsq.Message) error {
	msg.DisableAutoResponse()

	ctx := context.Background()

	payload, ok := parseNoticeMsg(msg.Body)
	if !ok {
		msg.Finish()
		return nil
	}

	notifyURL, secret, err := loadMerchant(ctx, p.db, payload.MerchantId)
	if err != nil || notifyURL == "" || secret == "" {
		msg.Finish()
		return nil
	}

	orderInfo, err := loadOrder(ctx, p.db, payload.OrderNo)
	if err != nil {
		msg.Finish()
		return nil
	}

	body, err := buildWebhookBody(orderInfo, secret)
	if err != nil {
		msg.Finish()
		return nil
	}

	statusCode, respBody, reqErr := p.postJSON(ctx, notifyURL, body)
	_ = insertNotifyLog(ctx, p.db, payload.MerchantId, payload.OrderNo, notifyURL, int(msg.Attempts), statusCode, string(respBody), reqErr)

	if reqErr == nil && statusCode >= 200 && statusCode < 300 {
		msg.Finish()
		return nil
	}

	attempt := int(msg.Attempts)
	if attempt <= len(p.delays) {
		msg.RequeueWithoutBackoff(p.delays[attempt-1])
		return nil
	}

	msg.Finish()
	return nil
}

func (p *Processor) postJSON(ctx context.Context, url string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil || resp == nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, p.maxRespBodyBytes))
	return resp.StatusCode, respBody, nil
}

type noticeMsg struct {
	MerchantId string `json:"merchant_id"`
	OrderNo    string `json:"order_no"`
}

func parseNoticeMsg(b []byte) (*noticeMsg, bool) {
	var payload noticeMsg
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, false
	}
	if payload.MerchantId == "" || payload.OrderNo == "" {
		return nil, false
	}
	return &payload, true
}

type orderRow struct {
	OrderNo         string
	MerchantId      string
	MerchantOrderNo string
	Amount          int64
	Currency        string
	Status          int32
	ChannelTradeNo  string
	PaidAmount      int64
}

func loadMerchant(ctx context.Context, db *gorm.DB, merchantId string) (string, string, error) {
	var r struct {
		NotifyURL string `gorm:"column:notify_url"`
		Secret    string `gorm:"column:app_secret"`
	}
	tx := db.WithContext(ctx).
		Table("merchants").
		Select("COALESCE(notify_url, '') AS notify_url, app_secret").
		Where("merchant_id = ? AND status = 1", merchantId).
		Limit(1).
		Take(&r)
	if tx.Error != nil {
		return "", "", tx.Error
	}
	return r.NotifyURL, r.Secret, nil
}

func loadOrder(ctx context.Context, db *gorm.DB, orderNo string) (*orderRow, error) {
	var o orderRow
	tx := db.WithContext(ctx).
		Table("payin_orders").
		Select("order_no, merchant_id, merchant_order_no, amount, currency, status, COALESCE(channel_trade_no,'') AS channel_trade_no, paid_amount").
		Where("order_no = ?", orderNo).
		Limit(1).
		Take(&o)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &o, nil
}

func buildWebhookBody(o *orderRow, secret string) ([]byte, error) {
	params := map[string]string{
		"order_no":          o.OrderNo,
		"merchant_id":       o.MerchantId,
		"merchant_order_no": o.MerchantOrderNo,
		"amount":            strconv.FormatInt(o.Amount, 10),
		"currency":          o.Currency,
		"status":            strconv.FormatInt(int64(o.Status), 10),
		"paid_amount":       strconv.FormatInt(o.PaidAmount, 10),
		"channel_trade_no": o.ChannelTradeNo,
	}
	sign := md5Sign(params, secret)
	out := map[string]any{
		"order_no":          o.OrderNo,
		"merchant_id":       o.MerchantId,
		"merchant_order_no": o.MerchantOrderNo,
		"amount":            o.Amount,
		"currency":          o.Currency,
		"status":            o.Status,
		"paid_amount":       o.PaidAmount,
		"channel_trade_no": o.ChannelTradeNo,
		"sign":              sign,
	}
	return json.Marshal(out)
}

func insertNotifyLog(ctx context.Context, db *gorm.DB, merchantId, orderNo, notifyUrl string, attempt int, httpStatus int, responseBody string, httpErr error) error {
	errMsg := ""
	if httpErr != nil {
		errMsg = httpErr.Error()
		if len(errMsg) > 255 {
			errMsg = errMsg[:255]
		}
	}
	if len(responseBody) > 8000 {
		responseBody = responseBody[:8000]
	}
	return db.WithContext(ctx).Exec(`
INSERT INTO merchant_notify_logs (merchant_id, order_no, notify_url, attempt, http_status, response_body, error_msg, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
`, merchantId, orderNo, notifyUrl, attempt, httpStatus, responseBody, errMsg).Error
}

func md5Sign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, strings.ToLower(k))
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
