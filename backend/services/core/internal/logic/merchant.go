package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gloopai/pay/common/consulx"
	"github.com/gloopai/pay/common/model"
	merchantpb "github.com/gloopai/pay/common/pb/merchant"
	"github.com/gloopai/pay/core/internal/kvcache"
	"github.com/gloopai/pay/core/internal/store"
	"github.com/gloopai/pay/core/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// 10 位十进制 merchant_id 上限（含）：与 fmt %010d 一致
const merchantNumericIDMax int64 = 9999999999

type CreateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMerchantLogic {
	return &CreateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMerchantLogic) CreateMerchant(in *merchantpb.CreateMerchantReq) (*merchantpb.UpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		floor, ferr := l.svcCtx.Merchants.GetMerchantNumericIDFloor(l.ctx)
		if ferr != nil {
			l.Errorf("read merchant_numeric_id_start: %v", ferr)
			floor = store.DefaultMerchantNumericIDFloor
		}
		n, err := l.svcCtx.Merchants.AllocNextMerchantNumericID(l.ctx, floor)
		if err != nil {
			l.Errorf("alloc merchant numeric id: %v", err)
			return nil, status.Error(codes.Internal, "allocate merchant id failed")
		}
		if n > merchantNumericIDMax {
			return nil, status.Error(codes.ResourceExhausted, "merchant id space exhausted")
		}
		merchantId = fmt.Sprintf("%010d", n)
	}
	appId := strings.TrimSpace(in.GetAppId())
	if appId == "" {
		appId = merchantId
	}
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email required")
	}
	passwordHash := strings.TrimSpace(in.GetPasswordHash())
	if passwordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "password_hash required")
	}

	secret := strings.TrimSpace(in.GetAppSecret())
	if secret == "" {
		v, err := newSecret()
		if err != nil {
			return nil, err
		}
		secret = v
	}

	statusVal := in.GetStatus()
	if statusVal == 0 {
		statusVal = 1
	}

	mc := strings.TrimSpace(in.GetMerchantConfig())
	if err := validateMerchantConfigJSON(mc); err != nil {
		return nil, err
	}
	rec := &model.Merchant{
		MerchantId:       merchantId,
		AppId:            appId,
		Email:            email,
		AppSecret:        secret,
		PasswordHash:     passwordHash,
		Status:           statusVal,
		IpWhitelist:      in.GetIpWhitelist(),
		NotifyUrl:        in.GetNotifyUrl(),
		ReturnUrl:        in.GetReturnUrl(),
		MerchantConfig:   mc,
		PayinBalance:     0,
		AvailableBalance: 0,
	}
	if err := l.svcCtx.Merchants.Create(l.ctx, rec); err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, status.Error(codes.AlreadyExists, "merchant already exists")
		}
		return nil, err
	}

	if len(in.GetPayinProductIds()) > 0 {
		grants := payinGrantsFromProductIDs(in.GetPayinProductIds())
		if err := l.svcCtx.MerchantPayinProducts.Replace(l.ctx, merchantId, grants); err != nil {
			return nil, status.Error(codes.Internal, "save merchant pay products failed")
		}
	}
	if len(in.GetPayoutProductIds()) > 0 {
		grants := payoutGrantsFromProductIDs(in.GetPayoutProductIds())
		if err := l.svcCtx.MerchantPayoutProducts.Replace(l.ctx, merchantId, grants); err != nil {
			return nil, status.Error(codes.Internal, "save merchant payout products failed")
		}
	}

	created, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Internal, "merchant create failed")
		}
		return nil, err
	}
	ids, _ := l.svcCtx.MerchantPayinProducts.ListProductIDs(l.ctx, merchantId)
	pids, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutProductIDs(l.ctx, merchantId)
	payinGrants, _ := l.svcCtx.MerchantPayinProducts.ListPayinGrants(l.ctx, merchantId)
	pg, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutGrants(l.ctx, merchantId)
	syncMerchantKV(l.ctx, l.svcCtx.RuntimeConfig, created)
	return &merchantpb.UpsertMerchantResp{Merchant: toMerchantInfo(created, ids, pids, payinGrants, pg)}, nil
}

func payinGrantsFromProductIDs(ids []int64) []model.PayinGrant {
	var out []model.PayinGrant
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		out = append(out, model.PayinGrant{PayinProductID: id, RateBps: nil})
	}
	return out
}

func payoutGrantsFromProductIDs(ids []int64) []model.PayoutGrant {
	var out []model.PayoutGrant
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		out = append(out, model.PayoutGrant{
			PayoutProductID: id,
			FeeMode:         1,
			RateBps:         nil,
			FixedFeeAmount:  0,
		})
	}
	return out
}

type UpdateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMerchantLogic {
	return &UpdateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMerchantLogic) UpdateMerchant(in *merchantpb.UpdateMerchantReq) (*merchantpb.UpsertMerchantResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}

	existing, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}

	secret := strings.TrimSpace(in.GetAppSecret())
	if secret == "" {
		secret = existing.AppSecret
	}
	passwordHash := strings.TrimSpace(in.GetPasswordHash())
	if passwordHash == "" {
		passwordHash = existing.PasswordHash
	}

	// status: 0 = 锁定，1 = 启用。须与请求一致落库；不可把 0 当成「未传」否则永远无法锁定。
	statusVal := in.GetStatus()

	merchantCfg := existing.MerchantConfig
	if in.MerchantConfig != nil {
		merchantCfg = strings.TrimSpace(*in.MerchantConfig)
		if err := validateMerchantConfigJSON(merchantCfg); err != nil {
			return nil, err
		}
	}
	rec := &model.Merchant{
		MerchantId:     merchantId,
		AppId:          existing.AppId,
		Email:          existing.Email,
		AppSecret:      secret,
		PasswordHash:   passwordHash,
		Status:         statusVal,
		IpWhitelist:    in.GetIpWhitelist(),
		NotifyUrl:      in.GetNotifyUrl(),
		ReturnUrl:      in.GetReturnUrl(),
		MerchantConfig: merchantCfg,
	}
	if err := l.svcCtx.Merchants.UpdateByMerchantId(l.ctx, merchantId, rec); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		return nil, err
	}
	ids, _ := l.svcCtx.MerchantPayinProducts.ListProductIDs(l.ctx, merchantId)
	pids, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutProductIDs(l.ctx, merchantId)
	payinGrants, _ := l.svcCtx.MerchantPayinProducts.ListPayinGrants(l.ctx, merchantId)
	pg, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutGrants(l.ctx, merchantId)
	syncMerchantKV(l.ctx, l.svcCtx.RuntimeConfig, updated)
	return &merchantpb.UpsertMerchantResp{Merchant: toMerchantInfo(updated, ids, pids, payinGrants, pg)}, nil
}

type GetMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMerchantLogic {
	return &GetMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMerchantLogic) GetMerchant(in *merchantpb.GetMerchantReq) (*merchantpb.GetMerchantResp, error) {
	merchantId := strings.TrimSpace(in.GetMerchantId())
	if merchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id required")
	}
	m, err := l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}
	ids, _ := l.svcCtx.MerchantPayinProducts.ListProductIDs(l.ctx, merchantId)
	pids, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutProductIDs(l.ctx, merchantId)
	payinGrants, _ := l.svcCtx.MerchantPayinProducts.ListPayinGrants(l.ctx, merchantId)
	pg, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutGrants(l.ctx, merchantId)
	return &merchantpb.GetMerchantResp{
		Merchant: toMerchantInfo(m, ids, pids, payinGrants, pg),
	}, nil
}

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
	appId := strings.TrimSpace(in.GetAppId())
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	var (
		m   *model.Merchant
		err error
	)
	switch {
	case merchantId != "":
		m, err = l.svcCtx.Merchants.GetByMerchantId(l.ctx, merchantId)
	case appId != "":
		m, err = l.svcCtx.Merchants.GetByAppId(l.ctx, appId)
	case email != "":
		m, err = l.svcCtx.Merchants.GetByEmail(l.ctx, email)
	default:
		return nil, status.Error(codes.InvalidArgument, "merchant_id or app_id or email required")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, err
	}
	cfgJSON := kvcache.PickMerchantConfig(l.svcCtx.MerchantSnapshot, m.MerchantId, m.MerchantConfig)
	baseSecret := m.AppSecret
	if snap, ok := l.svcCtx.MerchantSnapshot.Get(m.MerchantId); ok {
		baseSecret = snap.AppSecret
	}
	appSecret := appSecretFromMergedJSON(baseSecret, cfgJSON)
	return &merchantpb.GetAuthInfoResp{
		AppSecret:        appSecret,
		Status:           m.Status,
		IpWhitelist:      m.IpWhitelist,
		NotifyUrl:        m.NotifyUrl,
		ReturnUrl:        m.ReturnUrl,
		PayinBalance:     m.PayinBalance,
		AvailableBalance: m.AvailableBalance,
		MerchantId:       m.MerchantId,
		AppId:            m.AppId,
		Email:            m.Email,
		PasswordHash:     m.PasswordHash,
	}, nil
}

type ListMerchantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMerchantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMerchantsLogic {
	return &ListMerchantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMerchantsLogic) ListMerchants(in *merchantpb.ListMerchantsReq) (*merchantpb.ListMerchantsResp, error) {
	items, err := l.svcCtx.Merchants.List(l.ctx, in.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*merchantpb.MerchantInfo, 0, len(items))
	for i := range items {
		m := items[i]
		ids, err := l.svcCtx.MerchantPayinProducts.ListProductIDs(l.ctx, m.MerchantId)
		if err != nil {
			return nil, status.Error(codes.Internal, "load merchant pay products failed")
		}
		pids, err := l.svcCtx.MerchantPayoutProducts.ListPayoutProductIDs(l.ctx, m.MerchantId)
		if err != nil {
			return nil, status.Error(codes.Internal, "load merchant payout products failed")
		}
		payinGrants, _ := l.svcCtx.MerchantPayinProducts.ListPayinGrants(l.ctx, m.MerchantId)
		pg, _ := l.svcCtx.MerchantPayoutProducts.ListPayoutGrants(l.ctx, m.MerchantId)
		out = append(out, toMerchantInfo(&m, ids, pids, payinGrants, pg))
	}
	return &merchantpb.ListMerchantsResp{Merchants: out}, nil
}

func newSecret() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func toMerchantInfo(m *model.Merchant, payProductIds, payoutProductIds []int64, payinGrants []model.PayinGrant, pg []model.PayoutGrant) *merchantpb.MerchantInfo {
	if m == nil {
		return nil
	}
	pbCG := make([]*merchantpb.MerchantPayinGrant, 0, len(payinGrants))
	for _, g := range payinGrants {
		row := &merchantpb.MerchantPayinGrant{PayinProductId: g.PayinProductID}
		if g.RateBps != nil {
			v := *g.RateBps
			row.MerchantRateBps = &v
		}
		pbCG = append(pbCG, row)
	}
	pbPG := make([]*merchantpb.MerchantPayoutGrant, 0, len(pg))
	for _, g := range pg {
		mode := g.FeeMode
		if mode <= 0 {
			mode = 1
		}
		row := &merchantpb.MerchantPayoutGrant{
			PayoutProductId: g.PayoutProductID,
			FeeMode:         mode,
			FeeFixedAmount:  g.FixedFeeAmount,
		}
		if g.RateBps != nil {
			v := *g.RateBps
			row.MerchantRateBps = &v
		}
		pbPG = append(pbPG, row)
	}
	return &merchantpb.MerchantInfo{
		MerchantId:       m.MerchantId,
		AppId:            m.AppId,
		Email:            m.Email,
		AppSecret:        m.AppSecret,
		Status:           m.Status,
		IpWhitelist:      m.IpWhitelist,
		PayinBalance:     m.PayinBalance,
		AvailableBalance: m.AvailableBalance,
		FrozenBalance:    m.FrozenBalance,
		WithdrawnAmount:  m.WithdrawnAmount,
		NotifyUrl:        m.NotifyUrl,
		ReturnUrl:        m.ReturnUrl,
		PayinProductIds:  payProductIds,
		PayoutProductIds: payoutProductIds,
		PayinGrants:      pbCG,
		PayoutGrants:     pbPG,
		MerchantConfig:   m.MerchantConfig,
	}
}

func validateMerchantConfigJSON(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return status.Error(codes.InvalidArgument, "merchant_config must be valid JSON")
	}
	return nil
}

// appSecretFromMergedJSON returns app_secret from merged merchant_config JSON when non-empty;
// otherwise returns columnSecret (same idea as channel_config overriding sign_secret).
func appSecretFromMergedJSON(columnSecret, mergedJSON string) string {
	mergedJSON = strings.TrimSpace(mergedJSON)
	if mergedJSON == "" {
		return columnSecret
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(mergedJSON), &m); err != nil {
		return columnSecret
	}
	raw, ok := m["app_secret"]
	if !ok {
		return columnSecret
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return columnSecret
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return columnSecret
	}
	return s
}

func syncMerchantKV(ctx context.Context, store *consulx.ConfigStore, m *model.Merchant) {
	if store == nil || m == nil || strings.TrimSpace(m.MerchantId) == "" {
		return
	}
	key := consulx.MerchantSnapshotKVKey(m.MerchantId)
	if key == "" {
		return
	}
	kv := merchantKVFromStore(m)
	b, err := json.Marshal(kv)
	if err != nil {
		return
	}
	_ = store.PutBytes(ctx, key, b)
}

func merchantKVFromStore(m *model.Merchant) *model.MerchantKV {
	return &model.MerchantKV{
		ID:               m.ID,
		MerchantID:       m.MerchantId,
		AppID:            m.AppId,
		Email:            m.Email,
		AppSecret:        m.AppSecret,
		Status:           m.Status,
		IpWhitelist:      m.IpWhitelist,
		PayinBalance:     m.PayinBalance,
		AvailableBalance: m.AvailableBalance,
		FrozenBalance:    m.FrozenBalance,
		WithdrawnAmount:  m.WithdrawnAmount,
		NotifyURL:        m.NotifyUrl,
		ReturnURL:        m.ReturnUrl,
		MerchantConfig:   m.MerchantConfig,
	}
}
