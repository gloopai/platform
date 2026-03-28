package logic

import (
	"context"
	"strings"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
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
	if svcCtx.ChannelHub == nil {
		return nil, status.Error(codes.Internal, "channel hub not configured")
	}
	payinType := strings.TrimSpace(in.GetPayinType())
	amount := in.GetAmount()

	var cid, ppid int64
	var code string

	if payinType != "" {
		if merr := svcCtx.ChannelHub.MerchantPayinProductAllowed(ctx, merchantID, payinType); merr != nil {
			return nil, merr
		}
		var rerr error
		cid, ppid, rerr = svcCtx.ChannelHub.RoutePayin(ctx, payinType, amount)
		if rerr != nil {
			return nil, rerr
		}
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
