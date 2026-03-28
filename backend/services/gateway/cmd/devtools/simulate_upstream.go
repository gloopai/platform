package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func runSimulateUpstream(args []string) {
	fs := flag.NewFlagSet("simulate-upstream", flag.ExitOnError)
	baseURL := fs.String("base", "http://127.0.0.1:8090", "OpenAPI server base (POST /v1/callback/notify)")
	orderNo := fs.String("order_no", "", "platform order_no")
	paidAmount := fs.Int64("paid_amount", 0, "paid amount in cents")
	channelID := fs.Int64("channel_id", 0, "channel id")
	secret := fs.String("secret", "", "channel sign secret")
	tradeNo := fs.String("upstream_trade_no", "", "upstream trade no")
	_ = fs.Parse(args)

	if *orderNo == "" || *paidAmount <= 0 || *channelID <= 0 || *secret == "" {
		fmt.Fprintln(os.Stderr, "simulate-upstream: order_no, paid_amount, channel_id, secret required")
		fs.Usage()
		os.Exit(2)
	}
	tn := *tradeNo
	if tn == "" {
		tn = fmt.Sprintf("UP-%d", time.Now().UnixNano())
	}

	payload := map[string]any{
		"order_no":          *orderNo,
		"paid_amount":       *paidAmount,
		"upstream_trade_no": tn,
		"channel_id":        *channelID,
	}
	sign := md5Sign(map[string]string{
		"order_no":          *orderNo,
		"paid_amount":       strconv.FormatInt(*paidAmount, 10),
		"upstream_trade_no": tn,
		"channel_id":        strconv.FormatInt(*channelID, 10),
	}, *secret)
	payload["sign"] = sign

	b, _ := json.Marshal(payload)
	resp, err := http.Post(strings.TrimRight(*baseURL, "/")+"/v1/callback/notify", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "simulate-upstream: %v\n", err)
		os.Exit(2)
	}
	defer resp.Body.Close()
	fmt.Println("status:", resp.Status)
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
