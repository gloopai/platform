// Package mockpsp provides an in-memory mock PSP for tests and local integration.
//
// Usage:
//
//	d := mockpsp.New("")
//	reg := channeldriver.NewRegistry()
//	_ = mockpsp.RegisterAll(reg, d)
//	cfg := &channeldriver.ChannelConfig{DriverKey: mockpsp.DefaultDriverKey, SignSecret: "test_secret", AppID: "app1"}
//	resp, _ := d.CreatePayment(ctx, cfg, &channeldriver.CreatePaymentReq{MerchantOrderNo: "m1", AmountMinor: 100})
//	body, _ := mockpsp.BuildPayinNotifyBody(cfg, "m1", resp.ChannelOrderNo, channeldriver.PayinStatusSuccess, 100)
//	req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(body))
//	parsed, _ := d.VerifyPayinNotify(ctx, cfg, req)
package mockpsp
