package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type notifyPayload struct {
	OrderNo         string `json:"order_no"`
	MerchantID      string `json:"merchant_id"`
	MerchantOrderNo string `json:"merchant_order_no"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	Status          int32  `json:"status"`
	ChannelID       int64  `json:"channel_id"`
	PaidAmount      int64  `json:"paid_amount"`
	UpstreamTradeNo string `json:"upstream_trade_no"`
	Sign            string `json:"sign"`
}

func main() {
	listen := flag.String("listen", ":18090", "listen address")
	path := flag.String("path", "/notify", "callback path")
	secret := flag.String("secret", "demo_secret", "merchant api_secret used to verify sign")
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

		log.Printf("notify received at=%s order_no=%s merchant_id=%s status=%d paid_amount=%d channel_id=%d verify=%v", time.Now().Format(time.RFC3339), p.OrderNo, p.MerchantID, p.Status, p.PaidAmount, p.ChannelID, ok)
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
		"channel_id":        strconv.FormatInt(p.ChannelID, 10),
		"paid_amount":       strconv.FormatInt(p.PaidAmount, 10),
		"upstream_trade_no": p.UpstreamTradeNo,
	}
	return md5Sign(params, secret)
}

func md5Sign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		if b.Len() > 0 {
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
