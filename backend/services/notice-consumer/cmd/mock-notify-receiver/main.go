package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/pay/common/signmd5"
)

type notifyPayload struct {
	OrderNo         string `json:"order_no"`
	MerchantID      string `json:"merchant_id"`
	MerchantOrderNo string `json:"merchant_order_no"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	Status          int32  `json:"status"`
	PaidAmount      int64  `json:"paid_amount"`
	ChannelTradeNo string `json:"channel_trade_no"`
	Sign            string `json:"sign"`
}

func main() {
	listen := flag.String("listen", ":18090", "listen address")
	path := flag.String("path", "/notify", "callback path")
	secret := flag.String("secret", "demo_secret", "merchant app_secret used to verify sign")
	verify := flag.Bool("verify", true, "verify sign")
	flag.Parse()

	http.HandleFunc(*path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("method not allowed"))
			return
		}

		var p notifyPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid json"))
			return
		}

		ok := true
		if *verify {
			expect := signForPayload(p, *secret)
			ok = strings.EqualFold(expect, p.Sign)
		}

		log.Printf("notify received at=%s order_no=%s merchant_id=%s status=%d paid_amount=%d verify=%v", time.Now().Format(time.RFC3339), p.OrderNo, p.MerchantID, p.Status, p.PaidAmount, ok)
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Printf("payload:\n%s\n", string(b))

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad sign"))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	log.Printf("mock notify receiver listening on %s%s (verify=%v)", *listen, *path, *verify)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

func signForPayload(p notifyPayload, secret string) string {
	params := map[string]string{
		"order_no":          p.OrderNo,
		"merchant_id":       p.MerchantID,
		"merchant_order_no": p.MerchantOrderNo,
		"amount":            strconv.FormatInt(p.Amount, 10),
		"currency":          p.Currency,
		"status":            strconv.FormatInt(int64(p.Status), 10),
		"paid_amount":       strconv.FormatInt(p.PaidAmount, 10),
		"channel_trade_no": p.ChannelTradeNo,
	}
	return signmd5.SignSortedKV(params, secret)
}
