package portalnotify

import (
	"bytes"
	"sync"

	"github.com/gloopai/pay/common/notify"
	"github.com/zeromicro/go-zero/core/logx"
)

// Hub holds local SSE subscribers. NSQ delivers each message to every gateway instance; offline clients simply receive nothing.
type Hub struct {
	mu sync.RWMutex

	adminBroadcast []*subscriber // JWT admins: receive broadcast + personal
	adminMaster    []*subscriber   // master token: broadcast only
	adminByID      map[int64][]*subscriber

	merchantBroadcast []*subscriber
	merchantByID      map[string][]*subscriber
}

type subscriber struct {
	ch chan []byte
}

func NewHub() *Hub {
	return &Hub{
		adminByID:    make(map[int64][]*subscriber),
		merchantByID: make(map[string][]*subscriber),
	}
}

// RegisterAdmin registers an SSE connection. master=true => only broadcast; else adminID must be >0.
func (h *Hub) RegisterAdmin(master bool, adminID int64) (recv <-chan []byte, unregister func()) {
	s := &subscriber{ch: make(chan []byte, 64)}
	h.mu.Lock()
	if master {
		h.adminMaster = append(h.adminMaster, s)
	} else {
		h.adminBroadcast = append(h.adminBroadcast, s)
		h.adminByID[adminID] = append(h.adminByID[adminID], s)
	}
	h.mu.Unlock()
	return s.ch, func() { h.unregisterAdmin(s, master, adminID) }
}

func (h *Hub) unregisterAdmin(s *subscriber, master bool, adminID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if master {
		h.adminMaster = removeSub(h.adminMaster, s)
		return
	}
	h.adminBroadcast = removeSub(h.adminBroadcast, s)
	h.adminByID[adminID] = removeSub(h.adminByID[adminID], s)
	if len(h.adminByID[adminID]) == 0 {
		delete(h.adminByID, adminID)
	}
}

// RegisterMerchant registers a merchant SSE connection.
func (h *Hub) RegisterMerchant(merchantID string) (recv <-chan []byte, unregister func()) {
	s := &subscriber{ch: make(chan []byte, 64)}
	h.mu.Lock()
	h.merchantBroadcast = append(h.merchantBroadcast, s)
	h.merchantByID[merchantID] = append(h.merchantByID[merchantID], s)
	h.mu.Unlock()
	return s.ch, func() { h.unregisterMerchant(s, merchantID) }
}

func (h *Hub) unregisterMerchant(s *subscriber, merchantID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.merchantBroadcast = removeSub(h.merchantBroadcast, s)
	h.merchantByID[merchantID] = removeSub(h.merchantByID[merchantID], s)
	if len(h.merchantByID[merchantID]) == 0 {
		delete(h.merchantByID, merchantID)
	}
}

func removeSub(subs []*subscriber, s *subscriber) []*subscriber {
	out := subs[:0]
	for _, x := range subs {
		if x != s {
			out = append(out, x)
		}
	}
	return out
}

func (h *Hub) Dispatch(msg *notify.PortalNSQMessage) {
	if msg == nil {
		return
	}
	data := bytes.TrimSpace(msg.Envelope)
	if len(data) == 0 {
		return
	}
	switch msg.Portal {
	case "admin":
		if msg.Broadcast {
			h.broadcastAdmin(data)
		} else {
			for _, id := range msg.AdminUserIDs {
				if id > 0 {
					h.sendAdmin(id, data)
				}
			}
		}
	case "merchant":
		if msg.Broadcast {
			h.broadcastMerchant(data)
		} else {
			for _, mid := range msg.MerchantIDs {
				if mid != "" {
					h.sendMerchant(mid, data)
				}
			}
		}
	}
}

func (h *Hub) broadcastAdmin(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, s := range h.adminBroadcast {
		trySend(s, data)
	}
	for _, s := range h.adminMaster {
		trySend(s, data)
	}
}

func (h *Hub) sendAdmin(id int64, data []byte) {
	h.mu.RLock()
	subs := h.adminByID[id]
	if len(subs) == 0 {
		h.mu.RUnlock()
		logx.Infof("portal notify: no admin SSE subscribers for admin_id=%d (check JWT vs targeted ids)", id)
		return
	}
	defer h.mu.RUnlock()
	for _, s := range subs {
		trySend(s, data)
	}
}

func (h *Hub) broadcastMerchant(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, s := range h.merchantBroadcast {
		trySend(s, data)
	}
}

func (h *Hub) sendMerchant(id string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, s := range h.merchantByID[id] {
		trySend(s, data)
	}
}

func trySend(s *subscriber, data []byte) {
	if s == nil {
		return
	}
	payload := append([]byte(nil), data...)
	select {
	case s.ch <- payload:
	default:
		logx.Infof("portal notify: subscriber channel full, drop one message")
	}
}
