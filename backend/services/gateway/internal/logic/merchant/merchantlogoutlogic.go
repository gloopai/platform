package logic

import (
	"context"
	"strings"

	"github.com/gloopai/pay/gateway/internal/logic/shared"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MerchantLogoutLogic {
	return &MerchantLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantLogoutLogic) MerchantLogout(token string) (*types.MerchantLogoutResp, error) {
	tok := strings.TrimSpace(token)
	if tok != "" {
		_ = l.svcCtx.Sessions.DeleteMerchantSession(l.ctx, shared.TokenHash(tok))
	}
	return &types.MerchantLogoutResp{Ok: true}, nil
}
