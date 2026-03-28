package mockpsp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gloopai/pay/channeldriver"
	"github.com/gloopai/pay/channeldriver/mockpsp"
)

func TestMockPSP_PayinNotifyRoundTrip(t *testing.T) {
	ctx := context.Background()
	d := mockpsp.New("")
	cfg := &channeldriver.ChannelConfig{
		DriverKey:  mockpsp.DefaultDriverKey,
		SignSecret: "unit_test_secret",
		AppID:      "app_test",
	}
	resp, err := d.CreatePayment(ctx, cfg, &channeldriver.CreatePaymentReq{
		MerchantOrderNo: "M001",
		AmountMinor:     10_000,
		PayerName:       "n",
		PayerPhone:      "1",
		PayerEmail:      "e@e.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := mockpsp.BuildPayinNotifyBody(cfg, "M001", resp.ChannelOrderNo, channeldriver.PayinStatusSuccess, 10_000)
	if err != nil {
		t.Fatal(err)
	}
	req := mockpsp.NewJSONNotifyRequest("POST", "/notify", body)
	parsed, err := d.VerifyPayinNotify(ctx, cfg, req)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.MerchantOrderNo != "M001" || parsed.PaidAmountMinor != 10_000 {
		t.Fatalf("parsed %+v", parsed)
	}
	if string(d.PayinNotifyResponse(true)) != "SUCCESS" {
		t.Fatal("response")
	}
}

func TestMockPSP_Dispatch(t *testing.T) {
	ctx := context.Background()
	d := mockpsp.New("")
	reg := channeldriver.NewRegistry()
	if err := mockpsp.RegisterAll(reg, d); err != nil {
		t.Fatal(err)
	}
	cfg := &channeldriver.ChannelConfig{
		DriverKey:  mockpsp.DefaultDriverKey,
		SignSecret: "s",
	}
	resp, _ := d.CreatePayment(ctx, cfg, &channeldriver.CreatePaymentReq{MerchantOrderNo: "x", AmountMinor: 1})
	body, _ := mockpsp.BuildPayinNotifyBody(cfg, "x", resp.ChannelOrderNo, channeldriver.PayinStatusSuccess, 1)
	req := httptest.NewRequest("POST", "/cb", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	route := func(context.Context, *http.Request) (*channeldriver.ChannelConfig, error) {
		return cfg, nil
	}
	out, drv, err := channeldriver.HandlePayinNotify(ctx, reg, route, req, func(p *channeldriver.PayinNotifyParsed) (bool, error) {
		return p.Status == channeldriver.PayinStatusSuccess, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "SUCCESS" {
		t.Fatalf("got %q", out)
	}
	_ = drv
}
