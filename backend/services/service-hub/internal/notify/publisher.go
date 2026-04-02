package notify

import (
	"context"
	"strings"

	ntf "github.com/gloopai/platform/common/notify"
	"github.com/gloopai/platform/common/pb/servicehub"
	"github.com/gloopai/platform/service-hub/internal/store"
	"github.com/nsqio/go-nsq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Publisher struct {
	Producer *nsq.Producer
	Topic    string
	Store    *store.PortalNotificationsStore
}

func NewPublisher(producer *nsq.Producer, topic string, st *store.PortalNotificationsStore) *Publisher {
	return &Publisher{Producer: producer, Topic: topic, Store: st}
}

func (p *Publisher) Publish(ctx context.Context, req *servicehub.PublishPortalNotificationReq) (string, error) {
	if p == nil || p.Producer == nil {
		return "", status.Error(codes.FailedPrecondition, "notify nsq not configured")
	}
	topic := strings.TrimSpace(p.Topic)
	if topic == "" {
		topic = ntf.PortalNotifyTopic
	}

	res, err := ntf.Build(req)
	if err != nil {
		return "", ntf.GRPCStatusOrInternal(err)
	}

	row := persistRow(res.Row)
	if p.Store != nil && row != nil {
		if err := p.Store.Insert(ctx, row); err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}
	}

	if err := p.Producer.Publish(topic, res.NSQBody); err != nil {
		if p.Store != nil {
			_ = p.Store.DeleteByID(ctx, res.ID)
		}
		return "", status.Error(codes.Internal, err.Error())
	}
	return res.ID, nil
}

func persistRow(r *ntf.PersistRow) *store.PortalNotificationRow {
	if r == nil {
		return nil
	}
	return &store.PortalNotificationRow{
		ID:                    r.ID,
		Portal:                r.Portal,
		Broadcast:             r.Broadcast,
		Title:                 r.Title,
		Body:                  r.Body,
		Severity:              r.Severity,
		LinkPath:              r.LinkPath,
		LinkQueryJSON:         r.LinkQueryJSON,
		MetaJSON:              r.MetaJSON,
		TargetAdminIDsJSON:    r.TargetAdminIDsJSON,
		TargetMerchantIDsJSON: r.TargetMerchantIDsJSON,
		CreatedAt:             r.CreatedAt,
	}
}
