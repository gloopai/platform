package logic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gloopai/pay/gateway/internal/middleware"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MerchantRetryNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantRetryNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantRetryNotifyLogic {
	return &MerchantRetryNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantRetryNotifyLogic) MerchantRetryNotify(req *types.MerchantRetryNotifyReq) (*types.MerchantRetryNotifyResp, error) {
	merchantId := strings.TrimSpace(middleware.MerchantIdFromContext(l.ctx))
	if merchantId == "" || req.OrderNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no required")
	}
	body, _ := json.Marshal(map[string]any{
		"merchant_id": merchantId,
		"order_no":    req.OrderNo,
		"attempt":     0,
	})
	if err := l.svcCtx.NsqProducer.Publish(l.svcCtx.Config.Nsq.Topic, body); err != nil {
		return nil, err
	}
	return &types.MerchantRetryNotifyResp{Ok: true}, nil
}
