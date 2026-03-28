package server

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	channelpb "github.com/gloopai/pay/common/pb/channel"
	"github.com/gloopai/pay/core/internal/channelbridge/psp/contracts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ChannelServer) ChannelCreatePayment(ctx context.Context, req *channelpb.ChannelCreatePaymentReq) (*channelpb.ChannelCreatePaymentResp, error) {
	if req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	drv, err := s.svcCtx.GetChannelDriver(ctx, req.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	resp, err := drv.CreatePayment(ctx, &contracts.CreatePaymentReq{
		MerchantOrderNo: req.GetMerchantOrderNo(),
		AmountMinor:     req.GetAmountMinor(),
		PayerName:       req.GetPayerName(),
		PayerPhone:      req.GetPayerPhone(),
		PayerEmail:      req.GetPayerEmail(),
		UserIP:          req.GetUserIp(),
		NotifyURL:       req.GetNotifyUrl(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if resp == nil {
		return nil, status.Error(codes.Internal, "nil create payment response")
	}
	return &channelpb.ChannelCreatePaymentResp{
		PayUrl:         resp.PayURL,
		ChannelOrderNo: resp.ChannelOrderNo,
	}, nil
}

func (s *ChannelServer) ChannelVerifyPayinNotify(ctx context.Context, req *channelpb.ChannelVerifyPayinNotifyReq) (*channelpb.ChannelVerifyPayinNotifyResp, error) {
	if req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	drv, err := s.svcCtx.GetChannelDriver(ctx, req.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	ctOut := contracts.NotifyContentType(drv)
	failBody := drv.PayinNotifyResponse(false)
	method := strings.TrimSpace(req.GetMethod())
	if method == "" {
		method = http.MethodPost
	}
	r, err := http.NewRequestWithContext(ctx, method, "http://channel.local/notify/payin", bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	for k, v := range req.GetHeader() {
		r.Header.Set(k, v)
	}
	parsed, err := drv.VerifyPayinNotify(ctx, r)
	if err != nil {
		return &channelpb.ChannelVerifyPayinNotifyResp{
			VerifyOk:              false,
			ResponseBody:          failBody,
			ResponseContentType:   ctOut,
		}, nil
	}
	exp := strings.TrimSpace(req.GetExpectedOrderNo())
	if exp != "" && strings.TrimSpace(parsed.MerchantOrderNo) != exp {
		return &channelpb.ChannelVerifyPayinNotifyResp{
			VerifyOk:            false,
			ResponseBody:        failBody,
			ResponseContentType: ctOut,
		}, nil
	}
	return &channelpb.ChannelVerifyPayinNotifyResp{
		VerifyOk:            true,
		MerchantOrderNo:     parsed.MerchantOrderNo,
		ChannelOrderNo:      parsed.ChannelOrderNo,
		PaidAmountMinor:     parsed.PaidAmountMinor,
		PayinStatus:         int32(parsed.Status),
		RawStatus:           parsed.RawStatus,
		ResponseContentType: ctOut,
	}, nil
}

func (s *ChannelServer) ChannelBuildPayinNotifyResponse(ctx context.Context, req *channelpb.ChannelBuildPayinNotifyResponseReq) (*channelpb.ChannelBuildPayinNotifyResponseResp, error) {
	if req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	drv, err := s.svcCtx.GetChannelDriver(ctx, req.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &channelpb.ChannelBuildPayinNotifyResponseResp{
		Body:        drv.PayinNotifyResponse(req.GetSuccess()),
		ContentType: contracts.NotifyContentType(drv),
	}, nil
}

func (s *ChannelServer) ChannelVerifyPayoutNotify(ctx context.Context, req *channelpb.ChannelVerifyPayoutNotifyReq) (*channelpb.ChannelVerifyPayoutNotifyResp, error) {
	if req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	drv, err := s.svcCtx.GetChannelDriver(ctx, req.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	ctOut := contracts.NotifyContentType(drv)
	failBody := drv.PayoutNotifyResponse(false)
	method := strings.TrimSpace(req.GetMethod())
	if method == "" {
		method = http.MethodPost
	}
	r, err := http.NewRequestWithContext(ctx, method, "http://channel.local/notify/payout", bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	for k, v := range req.GetHeader() {
		r.Header.Set(k, v)
	}
	parsed, err := drv.VerifyPayoutNotify(ctx, r)
	if err != nil {
		return &channelpb.ChannelVerifyPayoutNotifyResp{
			VerifyOk:            false,
			ResponseBody:        failBody,
			ResponseContentType: ctOut,
		}, nil
	}
	exp := strings.TrimSpace(req.GetExpectedOrderNo())
	if exp != "" && strings.TrimSpace(parsed.MerchantOrderNo) != exp {
		return &channelpb.ChannelVerifyPayoutNotifyResp{
			VerifyOk:            false,
			ResponseBody:        failBody,
			ResponseContentType: ctOut,
		}, nil
	}
	return &channelpb.ChannelVerifyPayoutNotifyResp{
		VerifyOk:            true,
		MerchantOrderNo:     parsed.MerchantOrderNo,
		ChannelOrderNo:      parsed.ChannelOrderNo,
		AmountMinor:         parsed.AmountMinor,
		PayoutStatus:        int32(parsed.Status),
		ReferenceNo:         parsed.ReferenceNo,
		RawStatus:           parsed.RawStatus,
		ResponseContentType: ctOut,
	}, nil
}

func (s *ChannelServer) ChannelBuildPayoutNotifyResponse(ctx context.Context, req *channelpb.ChannelBuildPayoutNotifyResponseReq) (*channelpb.ChannelBuildPayoutNotifyResponseResp, error) {
	if req.GetChannelId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "channel_id required")
	}
	drv, err := s.svcCtx.GetChannelDriver(ctx, req.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &channelpb.ChannelBuildPayoutNotifyResponseResp{
		Body:        drv.PayoutNotifyResponse(req.GetSuccess()),
		ContentType: contracts.NotifyContentType(drv),
	}, nil
}
