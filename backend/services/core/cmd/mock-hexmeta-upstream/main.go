// Command mock-hexmeta-upstream is a minimal HTTP server that mimics the hexmeta-style upstream PSP API
// (paths under /exposed/v1, JSON envelope code/msg/data, MD5 sign). Use for local core/gateway tests.
//
// Run:
//
//	go run ./cmd/mock-hexmeta-upstream -listen :18088 -secret channel_secret_demo
//
// Point channel channel_config (or columns) to:
//   gateway_url: http://127.0.0.1:18088
//   app_id / channel_merchant_no: demo_app
//   sign_secret: channel_secret_demo
//
// The driver posts to: {gateway_url}/exposed/v1/order/payment  (and query/payment, makeup, order/payout, query/balance).
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

const apiPrefix = "/exposed/v1"

func gatewayURLHint(listen string) string {
	listen = strings.TrimSpace(listen)
	if listen == "" {
		return ""
	}
	if strings.HasPrefix(listen, ":") {
		return "http://127.0.0.1" + listen
	}
	return "http://" + listen
}

func main() {
	listen := flag.String("listen", ":18088", "listen address, e.g. :18088")
	secret := flag.String("secret", "channel_secret_demo", "must match channels.sign_secret / channel_config.sign_secret used by core")
	flag.Parse()

	s := &server{
		secret: *secret,
		orders: make(map[string]*orderRec),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(apiPrefix+"/order/payment", s.handlePayment)
	mux.HandleFunc(apiPrefix+"/query/payment", s.handleQueryPayment)
	mux.HandleFunc(apiPrefix+"/makeup", s.handleMakeup)
	mux.HandleFunc(apiPrefix+"/order/payout", s.handlePayout)
	mux.HandleFunc(apiPrefix+"/query/balance", s.handleBalance)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("mock hexmeta upstream listening on %s (secret=%q)", *listen, *secret)
	log.Printf("set channel gateway_url to %s", gatewayURLHint(*listen))
	log.Fatal(http.ListenAndServe(*listen, mux))
}

type server struct {
	mu     sync.Mutex
	secret string
	orders map[string]*orderRec // merchant orderNo -> state
	serial int
}

type orderRec struct {
	MerchantOrderNo string
	SysOrderNo      string
	Amount          string
	Status          string // "1" processing "2" success "3" failed
}

func (s *server) handlePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	m, err := readSignedJSON(r, s.secret)
	if err != nil {
		writeEnv(w, 0, err.Error(), nil)
		return
	}
	orderNo := strings.TrimSpace(m["orderNo"])
	if orderNo == "" {
		writeEnv(w, 0, "orderNo required", nil)
		return
	}
	amt := strings.TrimSpace(m["amount"])
	s.mu.Lock()
	s.serial++
	sys := fmt.Sprintf("MOCKSYS%d", s.serial)
	s.orders[orderNo] = &orderRec{
		MerchantOrderNo: orderNo,
		SysOrderNo:      sys,
		Amount:          amt,
		Status:          "1",
	}
	s.mu.Unlock()
	payURL := fmt.Sprintf("http://127.0.0.1/mock-pay?merchantOrderNo=%s&sysOrderNo=%s", orderNo, sys)
	writeEnv(w, 1, "ok", map[string]string{
		"sysOrderNo": sys,
		"payUrl":     payURL,
	})
}

func (s *server) handleQueryPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	m, err := readSignedJSON(r, s.secret)
	if err != nil {
		writeEnv(w, 0, err.Error(), nil)
		return
	}
	orderNo := strings.TrimSpace(m["orderNo"])
	s.mu.Lock()
	rec, ok := s.orders[orderNo]
	if ok && rec.Status == "1" {
		rec.Status = "2"
	}
	s.mu.Unlock()
	if !ok {
		writeEnv(w, 0, "order not found", nil)
		return
	}
	writeEnv(w, 1, "ok", map[string]string{
		"appId":       strings.TrimSpace(m["appId"]),
		"orderNo":     rec.MerchantOrderNo,
		"sysOrderNo":  rec.SysOrderNo,
		"amount":      rec.Amount,
		"status":      rec.Status,
		"referenceNo": "REF-" + rec.SysOrderNo,
		"failReason":  "",
	})
}

func (s *server) handleMakeup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, err := readSignedJSON(r, s.secret); err != nil {
		writeEnv(w, 0, err.Error(), nil)
		return
	}
	writeEnv(w, 1, "ok", map[string]string{})
}

func (s *server) handlePayout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	m, err := readSignedJSON(r, s.secret)
	if err != nil {
		writeEnv(w, 0, err.Error(), nil)
		return
	}
	_ = strings.TrimSpace(m["orderNo"])
	s.mu.Lock()
	s.serial++
	sys := fmt.Sprintf("MOCKPO%d", s.serial)
	s.mu.Unlock()
	writeEnv(w, 1, "ok", map[string]string{"sysOrderNo": sys})
}

func (s *server) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, err := readSignedJSON(r, s.secret); err != nil {
		writeEnv(w, 0, err.Error(), nil)
		return
	}
	writeEnv(w, 1, "ok", map[string]string{
		"availableBalance": "1000000",
		"unsettledAmount":  "0",
		"frozenAmount":     "0",
	})
}

func writeEnv(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// readSignedJSON parses POST JSON and verifies sign (same rules as hexmeta driver).
func readSignedJSON(r *http.Request, secret string) (map[string]string, error) {
	raw, err := readAll(r)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var rawMap map[string]interface{}
	if err := dec.Decode(&rawMap); err != nil || rawMap == nil {
		return nil, fmt.Errorf("invalid json")
	}
	sigRaw, ok := rawMap["sign"]
	if !ok {
		return nil, fmt.Errorf("missing sign")
	}
	sigStr := strings.TrimSpace(fmt.Sprint(sigRaw))
	if sigStr == "" {
		return nil, fmt.Errorf("empty sign")
	}
	params := make(map[string]string)
	for k, v := range rawMap {
		if k == "sign" {
			continue
		}
		s := valueForSign(v)
		if s == "" {
			continue
		}
		params[k] = s
	}
	if !strings.EqualFold(signMD5(params, secret), sigStr) {
		return nil, fmt.Errorf("bad sign")
	}
	out := make(map[string]string)
	for k, v := range rawMap {
		out[k] = valueForSign(v)
	}
	return out, nil
}

func readAll(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("empty body")
	}
	defer r.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

// signMD5 must match github.com/gloopai/pay/core/internal/channelbridge/psp/drivers/hexmeta signMD5.
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
