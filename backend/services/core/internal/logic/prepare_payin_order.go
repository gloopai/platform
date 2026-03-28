package logic

import (
	"context"
	"strings"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PreparePayinOrder 合并 OpenAPI 代收下单：白名单校验、选路、商户快照（计费用），避免网关多次 RPC。
func PreparePayinOrder(ctx context.Context, svcCtx *svc.ServiceContext, in *channelpb.PreparePayinOrderReq) (*channelpb.PreparePayinOrderResp, error) {
	merchantID := strings.TrimSpace(in.GetMerchantId())
	if merchantID == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	payinType := strings.TrimSpace(in.GetPayinType())
	amount := in.GetAmount()

	var cid, ppid int64
	var code string

	if payinType != "" {
		if svcCtx.OpenAPIMemoryReady() {
			ok := kvcache.MerchantHasPayinProductCodeMemory(
				merchantID,
				payinType,
				svcCtx.MerchantPayinGrantsSnapshot,
				svcCtx.PayinProductSnapshot,
			)
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
			}
		} else {
			ok, err := svcCtx.PayinProducts.MerchantHasPayinProductCode(ctx, merchantID, payinType)
			if err != nil {
				return nil, status.Error(codes.Internal, "check merchant pay products failed")
			}
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "payin_type not enabled for this merchant")
			}
		}

		routeL := NewRouteLogic(ctx, svcCtx)
		route, err := routeL.Route(&channelpb.RouteReq{Amount: amount, PayinType: payinType})
		if err != nil {
			return nil, err
		}
		cid = route.GetChannelId()
		ppid = route.GetPayinProductId()
		if cid <= 0 {
			return nil, status.Error(codes.FailedPrecondition, "no available channel for payin_type")
		}
		code = payinType
	}

	gm := NewGetMerchantLogic(ctx, svcCtx)
	mr, err := gm.GetMerchant(&merchantpb.GetMerchantReq{MerchantId: merchantID})
	if err != nil {
		return nil, err
	}
	return &channelpb.PreparePayinOrderResp{
		ChannelId:        cid,
		PayinProductId:   ppid,
		PayinProductCode: code,
		Merchant:         mr.GetMerchant(),
	}, nil
}
