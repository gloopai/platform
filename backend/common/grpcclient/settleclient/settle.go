package settleclient

import (
	"context"

	"github.com/gloopai/pay/common/pb/settle"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	CreditReq  = settle.CreditReq
	CreditResp = settle.CreditResp

	Settle interface {
		Credit(ctx context.Context, in *CreditReq, opts ...grpc.CallOption) (*CreditResp, error)
		ListFundLogs(ctx context.Context, in *settle.ListFundLogsReq, opts ...grpc.CallOption) (*settle.ListFundLogsResp, error)
	}

	defaultSettle struct {
		cli zrpc.Client
	}
)

func NewSettle(cli zrpc.Client) Settle {
	return &defaultSettle{cli: cli}
}

func (m *defaultSettle) client() settle.SettleClient {
	return settle.NewSettleClient(m.cli.Conn())
}

func (m *defaultSettle) Credit(ctx context.Context, in *CreditReq, opts ...grpc.CallOption) (*CreditResp, error) {
	return m.client().Credit(ctx, in, opts...)
}

func (m *defaultSettle) ListFundLogs(ctx context.Context, in *settle.ListFundLogsReq, opts ...grpc.CallOption) (*settle.ListFundLogsResp, error) {
	return m.client().ListFundLogs(ctx, in, opts...)
}
