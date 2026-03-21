package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
)

type noticeMsg struct {
	MerchantId string `json:"merchant_id"`
	OrderNo    string `json:"order_no"`
}

func main() {
	var (
		nsqdAddr     = flag.String("nsqd", "127.0.0.1:4150", "nsqd tcp addr")
		topic        = flag.String("topic", "merchant_notice", "nsq topic")
		channel      = flag.String("channel", "notice", "nsq channel")
		mysqlDSN     = flag.String("mysql_dsn", "root:your_password@tcp(127.0.0.1:3306)/pay?charset=utf8mb4&parseTime=true&loc=Local", "mysql dsn")
		timeout      = flag.Duration("timeout", 5*time.Second, "http timeout")
		consulAddr   = flag.String("consul_addr", "127.0.0.1:8500", "consul addr")
		consulSvc    = flag.String("consul_service", "notice-consumer", "consul service name")
		consulID     = flag.String("consul_id", "", "consul service id")
		consulHost   = flag.String("consul_host", "", "consul service host (optional)")
		healthListen = flag.String("health_listen", "0.0.0.0:8090", "health http listen addr")
	)
	flag.Parse()

	db, err := sql.Open("mysql", *mysqlDSN)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	httpClient := &http.Client{Timeout: *timeout}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	healthSrv := &http.Server{
		Addr:              *healthListen,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		ln, err := net.Listen("tcp", healthSrv.Addr)
		if err != nil {
			panic(err)
		}
		if err := healthSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	reg, err := registerConsul(*consulAddr, *consulSvc, *consulID, healthSrv.Addr, *consulHost)
	if err != nil {
		panic(err)
	}

	cfg := nsq.NewConfig()
	cfg.MaxAttempts = 6
	consumer, err := nsq.NewConsumer(*topic, *channel, cfg)
	if err != nil {
		panic(err)
	}
	delays := []time.Duration{15 * time.Second, 1 * time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour}

	consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {
		msg.DisableAutoResponse()

		var payload noticeMsg
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			msg.Finish()
			return nil
		}
		if payload.MerchantId == "" || payload.OrderNo == "" {
			msg.Finish()
			return nil
		}

		notifyURL, secret, err := loadMerchant(context.Background(), db, payload.MerchantId)
		if err != nil || notifyURL == "" || secret == "" {
			msg.Finish()
			return nil
		}

		orderInfo, err := loadOrder(context.Background(), db, payload.OrderNo)
		if err != nil {
			msg.Finish()
			return nil
		}

		body, err := buildWebhookBody(orderInfo, secret)
		if err != nil {
			msg.Finish()
			return nil
		}

		req, err := http.NewRequest(http.MethodPost, notifyURL, bytes.NewReader(body))
		if err != nil {
			msg.Finish()
			return nil
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		statusCode := 0
		var respBody []byte
		if err == nil && resp != nil {
			statusCode = resp.StatusCode
			respBody, _ = io.ReadAll(io.LimitReader(resp.Body, 8<<10))
			_ = resp.Body.Close()
		}
		_ = insertNotifyLog(context.Background(), db, payload.MerchantId, payload.OrderNo, notifyURL, int(msg.Attempts), statusCode, string(respBody), err)
		if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			msg.Finish()
			return nil
		}

		attempt := int(msg.Attempts)
		if attempt <= len(delays) {
			msg.RequeueWithoutBackoff(delays[attempt-1])
			return nil
		}
		msg.Finish()
		return nil
	}))

	if err := consumer.ConnectToNSQD(*nsqdAddr); err != nil {
		panic(err)
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()

	consumer.Stop()
	select {
	case <-consumer.StopChan:
	case <-time.After(3 * time.Second):
	}
	_ = reg.Deregister()
	_ = healthSrv.Shutdown(context.Background())
}

type orderRow struct {
	OrderNo         string
	MerchantId      string
	MerchantOrderNo string
	Amount          int64
	Currency        string
	Status          int32
	ChannelId       int64
	UpstreamTradeNo string
	PaidAmount      int64
}

func loadMerchant(ctx context.Context, db *sql.DB, merchantId string) (string, string, error) {
	var notifyURL, secret string
	if err := db.QueryRowContext(ctx, `
SELECT COALESCE(notify_url, ''), api_secret
FROM merchants
WHERE merchant_id = ? AND status = 1
LIMIT 1
`, merchantId).Scan(&notifyURL, &secret); err != nil {
		return "", "", err
	}
	return notifyURL, secret, nil
}

func loadOrder(ctx context.Context, db *sql.DB, orderNo string) (*orderRow, error) {
	var o orderRow
	if err := db.QueryRowContext(ctx, `
SELECT order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id, COALESCE(upstream_trade_no,''), paid_amount
FROM orders
WHERE order_no = ?
LIMIT 1
`, orderNo).Scan(&o.OrderNo, &o.MerchantId, &o.MerchantOrderNo, &o.Amount, &o.Currency, &o.Status, &o.ChannelId, &o.UpstreamTradeNo, &o.PaidAmount); err != nil {
		return nil, err
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
		"channel_id":        strconv.FormatInt(o.ChannelId, 10),
		"paid_amount":       strconv.FormatInt(o.PaidAmount, 10),
		"upstream_trade_no": o.UpstreamTradeNo,
	}
	sign := md5Sign(params, secret)
	out := map[string]any{
		"order_no":          o.OrderNo,
		"merchant_id":       o.MerchantId,
		"merchant_order_no": o.MerchantOrderNo,
		"amount":            o.Amount,
		"currency":          o.Currency,
		"status":            o.Status,
		"channel_id":        o.ChannelId,
		"paid_amount":       o.PaidAmount,
		"upstream_trade_no": o.UpstreamTradeNo,
		"sign":              sign,
	}
	return json.Marshal(out)
}

func insertNotifyLog(ctx context.Context, db *sql.DB, merchantId, orderNo, notifyUrl string, attempt int, httpStatus int, responseBody string, httpErr error) error {
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
	_, err := db.ExecContext(ctx, `
INSERT INTO merchant_notify_logs (merchant_id, order_no, notify_url, attempt, http_status, response_body, error_msg, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
`, merchantId, orderNo, notifyUrl, attempt, httpStatus, responseBody, errMsg)
	return err
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

type consulRegistrar struct {
	consulAddr string
	serviceID  string
	client     *http.Client
}

func registerConsul(consulAddr, serviceName, serviceID, listenOn, host string) (*consulRegistrar, error) {
	consulAddr = strings.TrimSpace(consulAddr)
	if consulAddr == "" {
		return nil, fmt.Errorf("consul addr required")
	}
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, fmt.Errorf("consul service name required")
	}

	lh, lp, err := net.SplitHostPort(listenOn)
	if err != nil {
		return nil, err
	}
	if host == "" || host == "0.0.0.0" {
		if lh != "" && lh != "0.0.0.0" {
			host = lh
		} else {
			host = "127.0.0.1"
		}
	}
	port, err := parsePort(lp)
	if err != nil {
		return nil, err
	}
	if serviceID == "" {
		serviceID = fmt.Sprintf("%s-%s-%d", serviceName, host, port)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	checkHost := host
	if host == "127.0.0.1" || host == "localhost" {
		nodeName := consulNodeName(client, consulAddr)
		if isLikelyDockerNodeName(nodeName) {
			checkHost = "host.docker.internal"
		}
	}

	payload := map[string]any{
		"Name":    serviceName,
		"ID":      serviceID,
		"Address": host,
		"Port":    port,
		"Check": map[string]any{
			"TCP":                            fmt.Sprintf("%s:%d", checkHost, port),
			"Interval":                       "10s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPut, "http://"+consulAddr+"/v1/agent/service/register", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("consul register failed: %s", resp.Status)
	}

	return &consulRegistrar{
		consulAddr: consulAddr,
		serviceID:  serviceID,
		client:     client,
	}, nil
}

func (r *consulRegistrar) Deregister() error {
	if r == nil || r.serviceID == "" || r.consulAddr == "" {
		return nil
	}
	req, err := http.NewRequest(http.MethodPut, "http://"+r.consulAddr+"/v1/agent/service/deregister/"+r.serviceID, nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func parsePort(s string) (int, error) {
	var p int
	_, err := fmt.Sscanf(s, "%d", &p)
	if err != nil {
		return 0, err
	}
	if p <= 0 || p > 65535 {
		return 0, fmt.Errorf("invalid port: %d", p)
	}
	return p, nil
}

func consulNodeName(client *http.Client, consulAddr string) string {
	req, err := http.NewRequest(http.MethodGet, "http://"+consulAddr+"/v1/agent/self", nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	var body struct {
		Config struct {
			NodeName string `json:"NodeName"`
		} `json:"Config"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body)
	return strings.TrimSpace(body.Config.NodeName)
}

func isLikelyDockerNodeName(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) != 12 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
