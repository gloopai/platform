package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gloopai/pay/common/signmd5"
)

func runSimulateChannelNotify(args []string) {
	fs := flag.NewFlagSet("simulate-channel", flag.ExitOnError)
	baseURL := fs.String("base", "http://127.0.0.1:8090", "OpenAPI server base (POST /v1/callback/notify)")
	orderNo := fs.String("order_no", "", "platform order_no")
	paidAmount := fs.Int64("paid_amount", 0, "paid amount in cents")
	channelID := fs.Int64("channel_id", 0, "channel id")
	secret := fs.String("secret", "", "channel sign secret")
	tradeNo := fs.String("channel_trade_no", "", "channel trade no")
	_ = fs.Parse(args)

	if *orderNo == "" || *paidAmount <= 0 || *channelID <= 0 || *secret == "" {
		fmt.Fprintln(os.Stderr, "simulate-channel: order_no, paid_amount, channel_id, secret required")
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
		"channel_trade_no": tn,
		"channel_id":        *channelID,
	}
	sign := signmd5.SignSortedKV(map[string]string{
		"order_no":          *orderNo,
		"paid_amount":       strconv.FormatInt(*paidAmount, 10),
		"channel_trade_no": tn,
		"channel_id":        strconv.FormatInt(*channelID, 10),
	}, *secret)
	payload["sign"] = sign

	b, _ := json.Marshal(payload)
	resp, err := http.Post(strings.TrimRight(*baseURL, "/")+"/v1/callback/notify", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "simulate-channel: %v\n", err)
		os.Exit(2)
	}
	defer resp.Body.Close()
	fmt.Println("status:", resp.Status)
}
