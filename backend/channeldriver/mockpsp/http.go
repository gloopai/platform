package mockpsp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/gloopai/pay/channeldriver"
)

// NewJSONNotifyRequest builds a POST request with JSON body for callback tests.
func NewJSONNotifyRequest(method, target string, body []byte) *http.Request {
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

// ChannelHTTPServer serves minimal /exposed/v1/* routes that delegate to Driver using fixed cfg.
// Intended for integration tests where GatewayBaseURL points at this server.
type ChannelHTTPServer struct {
	Cfg *channeldriver.ChannelConfig
	Drv *Driver
	srv *httptest.Server
}

// StartChannelHTTPServer starts httptest.Server; close with Close().
func StartChannelHTTPServer(cfg *channeldriver.ChannelConfig, drv *Driver) *ChannelHTTPServer {
	h := &ChannelHTTPServer{Cfg: cfg, Drv: drv}
	mux := http.NewServeMux()
	mux.HandleFunc("/exposed/v1/order/payment", h.handleCreatePayin)
	mux.HandleFunc("/exposed/v1/query/payment", h.handleQueryPayin)
	mux.HandleFunc("/exposed/v1/makeup", h.handleMakeup)
	mux.HandleFunc("/exposed/v1/order/payout", h.handleCreatePayout)
	mux.HandleFunc("/exposed/v1/query/balance", h.handleBalance)
	h.srv = httptest.NewServer(mux)
	return h
}

func (h *ChannelHTTPServer) BaseURL() string { return strings.TrimRight(h.srv.URL, "/") }

func (h *ChannelHTTPServer) Close() { h.srv.Close() }

type channelEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (h *ChannelHTTPServer) handleCreatePayin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		OrderNo   string `json:"orderNo"`
		Amount    string `json:"amount"`
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		UserIP    string `json:"userIP"`
		NotifyURL string `json:"notifyUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	amt, _ := strconv.ParseInt(body.Amount, 10, 64)
	ctx := r.Context()
	resp, err := h.Drv.CreatePayment(ctx, h.Cfg, &channeldriver.CreatePaymentReq{
		MerchantOrderNo: body.OrderNo,
		AmountMinor:     amt,
		PayerName:       body.Name,
		PayerPhone:      body.Phone,
		PayerEmail:      body.Email,
		UserIP:          body.UserIP,
		NotifyURL:       body.NotifyURL,
	})
	if err != nil {
		_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 0, Msg: err.Error()})
		return
	}
	data, _ := json.Marshal(map[string]string{
		"sysOrderNo": resp.ChannelOrderNo,
		"payUrl":     resp.PayURL,
	})
	_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 1, Msg: "OK", Data: data})
}

func (h *ChannelHTTPServer) handleQueryPayin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrderNo string `json:"orderNo"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	q, err := h.Drv.QueryPayment(r.Context(), h.Cfg, &channeldriver.QueryPaymentReq{MerchantOrderNo: body.OrderNo})
	if err != nil {
		_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 0, Msg: err.Error()})
		return
	}
	data, _ := json.Marshal(map[string]string{
		"appId":       h.Cfg.AppID,
		"orderNo":     q.MerchantOrderNo,
		"sysOrderNo":  q.ChannelOrderNo,
		"status":      q.RawStatus,
		"amount":      strconv.FormatInt(q.AmountMinor, 10),
		"referenceNo": q.ReferenceNo,
		"failReason":  q.FailReason,
	})
	_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 1, Msg: "OK", Data: data})
}

func (h *ChannelHTTPServer) handleMakeup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrderNo     string `json:"orderNo"`
		ReferenceNo string `json:"referenceNo"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	err := h.Drv.Makeup(r.Context(), h.Cfg, &channeldriver.MakeupReq{MerchantOrderNo: body.OrderNo, ReferenceNo: body.ReferenceNo})
	if err != nil {
		_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 0, Msg: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 1, Msg: "OK"})
}

func (h *ChannelHTTPServer) handleCreatePayout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrderNo   string `json:"orderNo"`
		WayCode   string `json:"wayCode"`
		Amount    string `json:"amount"`
		BankName  string `json:"bankName"`
		BankCode  string `json:"bankCode"`
		AccountNo string `json:"accountNo"`
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		NotifyURL string `json:"notifyUrl"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	amt, _ := strconv.ParseInt(body.Amount, 10, 64)
	wc := channeldriver.PayoutWayBankCard
	if body.WayCode == "2" {
		wc = channeldriver.PayoutWayUPI
	}
	resp, err := h.Drv.CreatePayout(r.Context(), h.Cfg, &channeldriver.CreatePayoutReq{
		MerchantOrderNo: body.OrderNo,
		AmountMinor:     amt,
		WayCode:         wc,
		BankName:        body.BankName,
		BankCode:        body.BankCode,
		AccountNo:       body.AccountNo,
		HolderName:      body.Name,
		Phone:           body.Phone,
		Email:           body.Email,
		NotifyURL:       body.NotifyURL,
	})
	if err != nil {
		_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 0, Msg: err.Error()})
		return
	}
	data, _ := json.Marshal(map[string]string{"sysOrderNo": resp.ChannelOrderNo})
	_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 1, Msg: "OK", Data: data})
}

func (h *ChannelHTTPServer) handleBalance(w http.ResponseWriter, r *http.Request) {
	bal, err := h.Drv.QueryBalance(r.Context(), h.Cfg)
	if err != nil {
		_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 0, Msg: err.Error()})
		return
	}
	data, _ := json.Marshal(map[string]string{
		"availableBalance": strconv.FormatInt(bal.AvailableMinor, 10),
		"unsettledAmount":  strconv.FormatInt(bal.UnsettledMinor, 10),
		"frozenAmount":     strconv.FormatInt(bal.FrozenMinor, 10),
	})
	_ = json.NewEncoder(w).Encode(channelEnvelope{Code: 1, Msg: "OK", Data: data})
}

