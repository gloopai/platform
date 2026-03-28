package base

import (
	"context"
	"net/http"
)

// PayinNotifyRoute resolves which driver_key and ChannelConfig apply to an incoming HTTP callback.
// The gateway implements this by path params (e.g. /notify/payin/:driver_key), query channel_id,
// or loading the order first to read channel_id from DB.
type PayinNotifyRoute func(ctx context.Context, r *http.Request) (cfg *ChannelConfig, err error)

// PayoutNotifyRoute is the payout counterpart of PayinNotifyRoute.
type PayoutNotifyRoute func(ctx context.Context, r *http.Request) (cfg *ChannelConfig, err error)

// HandlePayinNotify is a helper: resolve config, load driver by cfg.DriverKey, verify, return body bytes for the PSP.
// onSuccess is called with parsed notify when verification succeeds; if it returns false, PayinNotifyResponse(false) is used.
func HandlePayinNotify(ctx context.Context, reg *Registry, route PayinNotifyRoute, r *http.Request,
	onSuccess func(*PayinNotifyParsed) (ok bool, err error),
) (body []byte, drv PayinChannel, err error) {
	cfg, err := route(ctx, r)
	if err != nil {
		return nil, nil, err
	}
	drv, err = reg.Payin(cfg.DriverKey)
	if err != nil {
		return nil, nil, err
	}
	parsed, err := drv.VerifyPayinNotify(ctx, cfg, r)
	if err != nil {
		return drv.PayinNotifyResponse(false), drv, err
	}
	ok, err := onSuccess(parsed)
	if err != nil {
		return drv.PayinNotifyResponse(false), drv, err
	}
	if !ok {
		return drv.PayinNotifyResponse(false), drv, nil
	}
	return drv.PayinNotifyResponse(true), drv, nil
}

// HandlePayoutNotify is the payout analogue of HandlePayinNotify.
func HandlePayoutNotify(ctx context.Context, reg *Registry, route PayoutNotifyRoute, r *http.Request,
	onSuccess func(*PayoutNotifyParsed) (ok bool, err error),
) (body []byte, drv PayoutChannel, err error) {
	cfg, err := route(ctx, r)
	if err != nil {
		return nil, nil, err
	}
	drv, err = reg.Payout(cfg.DriverKey)
	if err != nil {
		return nil, nil, err
	}
	parsed, err := drv.VerifyPayoutNotify(ctx, cfg, r)
	if err != nil {
		return drv.PayoutNotifyResponse(false), drv, err
	}
	ok, err := onSuccess(parsed)
	if err != nil {
		return drv.PayoutNotifyResponse(false), drv, err
	}
	if !ok {
		return drv.PayoutNotifyResponse(false), drv, nil
	}
	return drv.PayoutNotifyResponse(true), drv, nil
}
