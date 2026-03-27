// 管理台：路由概览、统计大盘、全站订单
package handler

import (
	"net/http"

	"github.com/gloopai/pay/gateway/internal/apiresp"
	"github.com/gloopai/pay/gateway/internal/logic"
	"github.com/gloopai/pay/gateway/internal/svc"
	"github.com/gloopai/pay/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminRoutingSummaryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminRouting(r.Context(), svcCtx)
		resp, err := l.AdminRoutingSummary()
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminStatsOverviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminStats(r.Context(), svcCtx)
		resp, err := l.AdminStatsOverview()
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminListPayOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminOrdersReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminOrders(r.Context(), svcCtx)
		resp, err := l.AdminListPayOrders(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminListPayoutOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminOrdersReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminOrders(r.Context(), svcCtx)
		resp, err := l.AdminListPayoutOrders(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminMockPayoutSuccessHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminMockPayoutSuccessReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminOrders(r.Context(), svcCtx)
		resp, err := l.AdminMockPayoutSuccess(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminReconcileDayHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminDayOverviewReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminStats(r.Context(), svcCtx)
		resp, err := l.AdminDayOverview(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminSettlementLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminSettlementLogsReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSettlement(r.Context(), svcCtx)
		resp, err := l.AdminSettlementLogs(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminSettlementWithdrawalsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminSettlementWithdrawalsReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSettlement(r.Context(), svcCtx)
		resp, err := l.AdminSettlementWithdrawals(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminCreateWithdrawalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCreateWithdrawalReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSettlement(r.Context(), svcCtx)
		resp, err := l.AdminCreateWithdrawal(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminReviewWithdrawalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminReviewWithdrawalReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSettlement(r.Context(), svcCtx)
		resp, err := l.AdminReviewWithdrawal(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminPayoutWithdrawalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminPayoutWithdrawalReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSettlement(r.Context(), svcCtx)
		resp, err := l.AdminPayoutWithdrawal(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminSettlementDepositHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminDepositReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminSettlement(r.Context(), svcCtx)
		resp, err := l.AdminDeposit(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}

func AdminRefundsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminRefundsReq
		if err := httpx.Parse(r, &req); err != nil {
			apiresp.Fail(w, apiresp.CodeInvalidParams, err.Error())
			return
		}
		l := logic.NewAdminRefunds(r.Context(), svcCtx)
		resp, err := l.AdminListRefundCandidates(&req)
		if err != nil {
			apiresp.WriteFromGRPC(w, err)
		} else {
			apiresp.OK(w, resp)
		}
	}
}
