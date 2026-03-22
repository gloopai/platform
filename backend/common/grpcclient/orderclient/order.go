package orderclient

import (
	"context"

	"github.com/gloopai/pay/common/pb/order"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	CreateOrderReq   = order.CreateOrderReq
	CreateOrderResp  = order.CreateOrderResp
	GetOrderReq      = order.GetOrderReq
	GetOrderResp     = order.GetOrderResp
	MarkPaidReq      = order.MarkPaidReq
	MarkPaidResp     = order.MarkPaidResp
	ListOrdersReq    = order.ListOrdersReq
	ListOrdersResp   = order.ListOrdersResp
	TodaySummaryReq  = order.TodaySummaryReq
	TodaySummaryResp = order.TodaySummaryResp
	PrepareTerminalPayReq  = order.PrepareTerminalPayReq
	PrepareTerminalPayResp = order.PrepareTerminalPayResp
	OrderInfo              = order.OrderInfo

	Order interface {
		CreateOrder(ctx context.Context, in *CreateOrderReq, opts ...grpc.CallOption) (*CreateOrderResp, error)
		GetOrder(ctx context.Context, in *GetOrderReq, opts ...grpc.CallOption) (*GetOrderResp, error)
		MarkPaid(ctx context.Context, in *MarkPaidReq, opts ...grpc.CallOption) (*MarkPaidResp, error)
		ListOrders(ctx context.Context, in *ListOrdersReq, opts ...grpc.CallOption) (*ListOrdersResp, error)
		TodaySummary(ctx context.Context, in *TodaySummaryReq, opts ...grpc.CallOption) (*TodaySummaryResp, error)
		PrepareTerminalPay(ctx context.Context, in *PrepareTerminalPayReq, opts ...grpc.CallOption) (*PrepareTerminalPayResp, error)
		AdminTodayOverview(ctx context.Context, in *order.AdminTodayOverviewReq, opts ...grpc.CallOption) (*order.AdminTodayOverviewResp, error)
		ListMerchantNotifyLogs(ctx context.Context, in *order.ListMerchantNotifyLogsReq, opts ...grpc.CallOption) (*order.ListMerchantNotifyLogsResp, error)
	}

	defaultOrder struct {
		cli zrpc.Client
	}
)

func NewOrder(cli zrpc.Client) Order {
	return &defaultOrder{cli: cli}
}

func (m *defaultOrder) client() order.OrderClient {
	return order.NewOrderClient(m.cli.Conn())
}

func (m *defaultOrder) CreateOrder(ctx context.Context, in *CreateOrderReq, opts ...grpc.CallOption) (*CreateOrderResp, error) {
	return m.client().CreateOrder(ctx, in, opts...)
}

func (m *defaultOrder) GetOrder(ctx context.Context, in *GetOrderReq, opts ...grpc.CallOption) (*GetOrderResp, error) {
	return m.client().GetOrder(ctx, in, opts...)
}

func (m *defaultOrder) MarkPaid(ctx context.Context, in *MarkPaidReq, opts ...grpc.CallOption) (*MarkPaidResp, error) {
	return m.client().MarkPaid(ctx, in, opts...)
}

func (m *defaultOrder) ListOrders(ctx context.Context, in *ListOrdersReq, opts ...grpc.CallOption) (*ListOrdersResp, error) {
	return m.client().ListOrders(ctx, in, opts...)
}

func (m *defaultOrder) TodaySummary(ctx context.Context, in *TodaySummaryReq, opts ...grpc.CallOption) (*TodaySummaryResp, error) {
	return m.client().TodaySummary(ctx, in, opts...)
}

func (m *defaultOrder) PrepareTerminalPay(ctx context.Context, in *PrepareTerminalPayReq, opts ...grpc.CallOption) (*PrepareTerminalPayResp, error) {
	return m.client().PrepareTerminalPay(ctx, in, opts...)
}

func (m *defaultOrder) AdminTodayOverview(ctx context.Context, in *order.AdminTodayOverviewReq, opts ...grpc.CallOption) (*order.AdminTodayOverviewResp, error) {
	return m.client().AdminTodayOverview(ctx, in, opts...)
}

func (m *defaultOrder) ListMerchantNotifyLogs(ctx context.Context, in *order.ListMerchantNotifyLogsReq, opts ...grpc.CallOption) (*order.ListMerchantNotifyLogsResp, error) {
	return m.client().ListMerchantNotifyLogs(ctx, in, opts...)
}
