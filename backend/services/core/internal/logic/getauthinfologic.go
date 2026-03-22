package logic

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gloopai/pay/core/internal/svc"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetAuthInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuthInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthInfoLogic {
	return &GetAuthInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuthInfoLogic) GetAuthInfo(in *merchantpb.GetAuthInfoReq) (*merchantpb.GetAuthInfoResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	m, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}
	return &merchantpb.GetAuthInfoResp{
		ApiSecret:   m.ApiSecret,
		Status:      m.Status,
		IpWhitelist: m.IpWhitelist,
		NotifyUrl:   m.NotifyUrl,
		ReturnUrl:   m.ReturnUrl,
		Balance:     m.Balance,
		RateBps:     m.RateBps,
	}, nil
}
