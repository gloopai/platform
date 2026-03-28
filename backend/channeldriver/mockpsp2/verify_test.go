package mockpsp2

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gloopai/pay/channeldriver"
)

func TestVerifyPayinNotify_roundtrip(t *testing.T) {
	cfg := &channeldriver.ChannelConfig{SignSecret: "channel_secret_alt"}
	body, err := BuildPayinNotifyBody(cfg, "P-ORD-1", "ALT99", channeldriver.PayinStatusSuccess, 100)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/cb", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	d := New("")
	p, err := d.VerifyPayinNotify(context.Background(), cfg, req)
	if err != nil {
		t.Fatal(err)
	}
	if p.MerchantOrderNo != "P-ORD-1" || p.UpstreamOrderNo != "ALT99" || p.PaidAmountMinor != 100 || p.Status != channeldriver.PayinStatusSuccess {
		t.Fatalf("parsed=%+v", p)
	}
}

func TestVerifyPayoutNotify_roundtrip(t *testing.T) {
	cfg := &channeldriver.ChannelConfig{SignSecret: "channel_secret_alt"}
	body, err := BuildPayoutNotifyBody(cfg, "PO-1", "ALTPO9", channeldriver.PayoutStatusSuccess, 200, "UTR-X")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/cb", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	d := New("")
	p, err := d.VerifyPayoutNotify(context.Background(), cfg, req)
	if err != nil {
		t.Fatal(err)
	}
	if p.MerchantOrderNo != "PO-1" || p.UpstreamOrderNo != "ALTPO9" || p.AmountMinor != 200 || p.Status != channeldriver.PayoutStatusSuccess {
		t.Fatalf("parsed=%+v", p)
	}
}
