package portalnotify

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gloopai/pay/common/notify"
	"github.com/gloopai/pay/gateway/internal/config"
	"github.com/nsqio/go-nsq"
	"github.com/zeromicro/go-zero/core/logx"
)

// StartConsumer runs an NSQ consumer (unique channel per process so each gateway receives every message).
// Returns a stop function; no-op if nsqd address is empty.
func StartConsumer(c config.Config, hub *Hub) (stop func()) {
	if hub == nil {
		return func() {}
	}
	addr := strings.TrimSpace(c.Nsq.NsqdTCPAddr)
	if addr == "" {
		logx.Info("portal notify: NSQ disabled (Nsq.NsqdTCPAddr empty), SSE will not receive pushes")
		return func() {}
	}
	topic := strings.TrimSpace(c.Nsq.PortalNotifyTopic)
	if topic == "" {
		topic = notify.PortalNotifyTopic
	}
	hostname, _ := os.Hostname()
	channel := fmt.Sprintf("sse_%s_%d_%d", sanitizeChannel(hostname), os.Getpid(), os.Getppid())
	if len(channel) > 96 {
		channel = channel[:96]
	}

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = 64

	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		panic(err)
	}
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var wire notify.PortalNSQMessage
		if err := json.Unmarshal(message.Body, &wire); err != nil {
			logx.Errorf("portal notify: bad nsq body: %v", err)
			return nil
		}
		hub.Dispatch(&wire)
		return nil
	}))
	if err := consumer.ConnectToNSQD(addr); err != nil {
		panic(err)
	}
	logx.Infof("portal notify: NSQ consumer topic=%s channel=%s nsqd=%s", topic, channel, addr)
	return func() {
		consumer.Stop()
	}
}

func sanitizeChannel(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b = append(b, r)
		default:
			b = append(b, '_')
		}
	}
	if len(b) > 32 {
		b = b[:32]
	}
	if len(b) == 0 {
		return "gw"
	}
	return string(b)
}
