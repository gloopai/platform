package settleclient

import (
	"context"

	"github.com/gloopai/pay/common/pb/settle"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	CreditReq                 = settle.CreditReq
	CreditResp                = settle.CreditResp
	DebitPayoutReq            = settle.DebitPayoutReq
	DebitPayoutResp           = settle.DebitPayoutResp
	TransferPayinToPayoutReq  = settle.TransferPayinToPayoutReq
	TransferPayinToPayoutResp = settle.TransferPayinToPayoutResp
	CreateWithdrawalReq       = settle.CreateWithdrawalReq
	CreateWithdrawalResp      = settle.CreateWithdrawalResp
	ListWithdrawalsReq        = settle.ListWithdrawalsReq
	ListWithdrawalsResp       = settle.ListWithdrawalsResp
	ReviewWithdrawalReq       = settle.ReviewWithdrawalReq
	ReviewWithdrawalResp      = settle.ReviewWithdrawalResp
	PayoutWithdrawalReq       = settle.PayoutWithdrawalReq
	PayoutWithdrawalResp      = settle.PayoutWithdrawalResp

	Settle interface {
		Credit(ctx context.Context, in *CreditReq, opts ...grpc.CallOption) (*CreditResp, error)
		DebitPayout(ctx context.Context, in *DebitPayoutReq, opts ...grpc.CallOption) (*DebitPayoutResp, error)
		TransferPayinToPayout(ctx context.Context, in *TransferPayinToPayoutReq, opts ...grpc.CallOption) (*TransferPayinToPayoutResp, error)
		ListFundLogs(ctx context.Context, in *settle.ListFundLogsReq, opts ...grpc.CallOption) (*settle.ListFundLogsResp, error)
		CreateWithdrawal(ctx context.Context, in *CreateWithdrawalReq, opts ...grpc.CallOption) (*CreateWithdrawalResp, error)
		ListWithdrawals(ctx context.Context, in *ListWithdrawalsReq, opts ...grpc.CallOption) (*ListWithdrawalsResp, error)
		ReviewWithdrawal(ctx context.Context, in *ReviewWithdrawalReq, opts ...grpc.CallOption) (*ReviewWithdrawalResp, error)
		PayoutWithdrawal(ctx context.Context, in *PayoutWithdrawalReq, opts ...grpc.CallOption) (*PayoutWithdrawalResp, error)
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

func (m *defaultSettle) DebitPayout(ctx context.Context, in *DebitPayoutReq, opts ...grpc.CallOption) (*DebitPayoutResp, error) {
	return m.client().DebitPayout(ctx, in, opts...)
}

func (m *defaultSettle) TransferPayinToPayout(ctx context.Context, in *TransferPayinToPayoutReq, opts ...grpc.CallOption) (*TransferPayinToPayoutResp, error) {
	return m.client().TransferPayinToPayout(ctx, in, opts...)
}

func (m *defaultSettle) ListFundLogs(ctx context.Context, in *settle.ListFundLogsReq, opts ...grpc.CallOption) (*settle.ListFundLogsResp, error) {
	return m.client().ListFundLogs(ctx, in, opts...)
}

func (m *defaultSettle) CreateWithdrawal(ctx context.Context, in *CreateWithdrawalReq, opts ...grpc.CallOption) (*CreateWithdrawalResp, error) {
	return m.client().CreateWithdrawal(ctx, in, opts...)
}

func (m *defaultSettle) ListWithdrawals(ctx context.Context, in *ListWithdrawalsReq, opts ...grpc.CallOption) (*ListWithdrawalsResp, error) {
	return m.client().ListWithdrawals(ctx, in, opts...)
}

func (m *defaultSettle) ReviewWithdrawal(ctx context.Context, in *ReviewWithdrawalReq, opts ...grpc.CallOption) (*ReviewWithdrawalResp, error) {
	return m.client().ReviewWithdrawal(ctx, in, opts...)
}

func (m *defaultSettle) PayoutWithdrawal(ctx context.Context, in *PayoutWithdrawalReq, opts ...grpc.CallOption) (*PayoutWithdrawalResp, error) {
	return m.client().PayoutWithdrawal(ctx, in, opts...)
}
