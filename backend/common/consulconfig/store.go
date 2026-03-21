package consulconfig

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/consul"
	"github.com/hashicorp/consul/api"
)

type Event struct {
	Key  string
	Data []byte
}

type Store struct {
	client   *api.Client
	prefixes []string

	mu    sync.RWMutex
	data  map[string][]byte
	subs  map[chan Event]struct{}
	stop  chan struct{}
	close sync.Once
}

func GlobalPrefix() string {
	return "pay/config/global/"
}

func ServicePrefix(serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "pay/config/services/unknown/"
	}
	return "pay/config/services/" + serviceName + "/"
}

func NewStore(consulAddr string, prefixes ...string) (*Store, error) {
	client, err := consul.NewClient(consulAddr)
	if err != nil {
		return nil, err
	}
	var ps []string
	for _, p := range prefixes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		ps = append(ps, p)
	}
	if len(ps) == 0 {
		return nil, errors.New("at least one prefix required")
	}
	s := &Store{
		client:   client,
		prefixes: ps,
		data:     make(map[string][]byte),
		subs:     make(map[chan Event]struct{}),
		stop:     make(chan struct{}),
	}
	return s, nil
}

func (s *Store) Start() {
	for _, p := range s.prefixes {
		prefix := p
		go s.watchPrefix(prefix)
	}
}

func (s *Store) Stop() {
	s.close.Do(func() {
		close(s.stop)
		s.mu.Lock()
		for ch := range s.subs {
			close(ch)
		}
		s.subs = map[chan Event]struct{}{}
		s.mu.Unlock()
	})
}

func (s *Store) Subscribe(buffer int) <-chan Event {
	if buffer <= 0 {
		buffer = 16
	}
	ch := make(chan Event, buffer)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *Store) GetBytes(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil, false
	}
	out := make([]byte, len(v))
	copy(out, v)
	return out, true
}

func (s *Store) GetJSON(key string, out any) bool {
	b, ok := s.GetBytes(key)
	if !ok {
		return false
	}
	return json.Unmarshal(b, out) == nil
}

func (s *Store) PutBytes(ctx context.Context, key string, value []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key required")
	}
	pair := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := s.client.KV().Put(pair, nilWithContext(ctx))
	return err
}

func (s *Store) PutJSON(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.PutBytes(ctx, key, b)
}

func (s *Store) Delete(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key required")
	}
	_, err := s.client.KV().Delete(key, nilWithContext(ctx))
	return err
}

func (s *Store) watchPrefix(prefix string) {
	var (
		lastIndex uint64
		backoff   = 200 * time.Millisecond
	)

	for {
		select {
		case <-s.stop:
			return
		default:
		}

		nextIndex, err := s.pullPrefix(prefix, lastIndex)
		if err != nil {
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
			case <-s.stop:
				timer.Stop()
				return
			}
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = 200 * time.Millisecond
		if nextIndex > lastIndex {
			lastIndex = nextIndex
		}
	}
}

func (s *Store) pullPrefix(prefix string, lastIndex uint64) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 310*time.Second)
	defer cancel()

	opts := (&api.QueryOptions{
		WaitIndex: lastIndex,
		WaitTime:  300 * time.Second,
	}).WithContext(ctx)

	pairs, meta, err := s.client.KV().List(prefix, opts)
	if err != nil {
		return lastIndex, err
	}

	updates := make(map[string][]byte)
	for _, p := range pairs {
		if p == nil {
			continue
		}
		key := strings.TrimSpace(p.Key)
		if key == "" {
			continue
		}
		val := make([]byte, len(p.Value))
		copy(val, p.Value)
		updates[key] = val
	}

	var events []Event
	s.mu.Lock()
	for k, v := range updates {
		old, ok := s.data[k]
		if !ok || !bytesEqual(old, v) {
			s.data[k] = v
			events = append(events, Event{Key: k, Data: v})
		}
	}
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			if _, ok := updates[k]; !ok {
				delete(s.data, k)
				events = append(events, Event{Key: k, Data: nil})
			}
		}
	}
	for ch := range s.subs {
		for i := range events {
			select {
			case ch <- events[i]:
			default:
			}
		}
	}
	s.mu.Unlock()

	if meta == nil || meta.LastIndex == 0 {
		return lastIndex, nil
	}
	return meta.LastIndex, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func nilWithContext(ctx context.Context) *api.WriteOptions {
	if ctx == nil {
		return nil
	}
	return (&api.WriteOptions{}).WithContext(ctx)
}
