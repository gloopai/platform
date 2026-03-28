package consulx

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gloopai/pay/common/configkv"
	"github.com/hashicorp/consul/api"
)

type ConfigEvent struct {
	Key  string
	Data []byte
}

type ConfigStore struct {
	client   *api.Client
	prefixes []string

	mu            sync.RWMutex
	data          map[string][]byte
	subs          map[chan ConfigEvent]struct{}
	watchedKeys   map[string]struct{}
	nextHandlerID int
	handlers      map[string]map[int]func(ConfigEvent)
	stop          chan struct{}
	close         sync.Once
}

func NewConfigStore(consulAddr string, prefixes ...string) (*ConfigStore, error) {
	client, err := NewClient(consulAddr)
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
	s := &ConfigStore{
		client:      client,
		prefixes:    ps,
		data:        make(map[string][]byte),
		subs:        make(map[chan ConfigEvent]struct{}),
		watchedKeys: make(map[string]struct{}),
		handlers:    make(map[string]map[int]func(ConfigEvent)),
		stop:        make(chan struct{}),
	}
	return s, nil
}

func NewDefaultConfigStore(serviceName string) (*ConfigStore, error) {
	return NewConfigStore("", configkv.GlobalConfigPrefix(), configkv.ServiceConfigPrefix(serviceName))
}

func (s *ConfigStore) Start() {
	for _, p := range s.prefixes {
		prefix := p
		go s.watchPrefix(prefix)
	}
}

func (s *ConfigStore) Stop() {
	s.close.Do(func() {
		close(s.stop)
		s.mu.Lock()
		for ch := range s.subs {
			close(ch)
		}
		s.subs = map[chan ConfigEvent]struct{}{}
		s.mu.Unlock()
	})
}

func (s *ConfigStore) Subscribe(buffer int) <-chan ConfigEvent {
	if buffer <= 0 {
		buffer = 16
	}
	ch := make(chan ConfigEvent, buffer)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *ConfigStore) OnKey(key string, fn func(ConfigEvent)) (cancel func()) {
	key = strings.TrimSpace(key)
	if key == "" || fn == nil {
		return func() {}
	}
	s.mu.Lock()
	s.nextHandlerID++
	id := s.nextHandlerID
	m := s.handlers[key]
	if m == nil {
		m = make(map[int]func(ConfigEvent))
		s.handlers[key] = m
	}
	m[id] = fn
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		if m := s.handlers[key]; m != nil {
			delete(m, id)
			if len(m) == 0 {
				delete(s.handlers, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *ConfigStore) WatchKey(key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	s.mu.Lock()
	if _, ok := s.watchedKeys[key]; ok {
		s.mu.Unlock()
		return
	}
	s.watchedKeys[key] = struct{}{}
	s.mu.Unlock()

	go s.watchKey(key)
}

// SyncPrefixOnce performs one blocking List(prefix) (same as the first long-poll in watchPrefix with index 0)
// and merges results into the in-memory cache. Use after Start() to avoid racing the initial watch goroutine.
func (s *ConfigStore) SyncPrefixOnce(ctx context.Context, prefix string) error {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return errors.New("prefix required")
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	_, err := s.pullPrefix(prefix, 0)
	return err
}

// ForEachPrefix invokes fn for every cached key under prefix (read lock). Prefix should end with "/".
func (s *ConfigStore) ForEachPrefix(prefix string, fn func(key string, data []byte)) {
	if fn == nil {
		return
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.data {
		if strings.HasPrefix(k, prefix) {
			out := make([]byte, len(v))
			copy(out, v)
			fn(k, out)
		}
	}
}

func (s *ConfigStore) GetBytes(key string) ([]byte, bool) {
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

func (s *ConfigStore) GetJSON(key string, out any) bool {
	b, ok := s.GetBytes(key)
	if !ok {
		return false
	}
	return json.Unmarshal(b, out) == nil
}

func (s *ConfigStore) PutBytes(ctx context.Context, key string, value []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key required")
	}
	pair := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := s.client.KV().Put(pair, nilWriteOptions(ctx))
	if err != nil {
		return err
	}

	val := make([]byte, len(value))
	copy(val, value)
	s.updateKey(key, val)
	return nil
}

func (s *ConfigStore) PutJSON(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.PutBytes(ctx, key, b)
}

func (s *ConfigStore) Delete(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key required")
	}
	_, err := s.client.KV().Delete(key, nilWriteOptions(ctx))
	if err != nil {
		return err
	}
	s.updateKey(key, nil)
	return nil
}

func (s *ConfigStore) Fetch(ctx context.Context, key string) ([]byte, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, false, errors.New("key required")
	}
	var opts *api.QueryOptions
	if ctx != nil {
		opts = (&api.QueryOptions{}).WithContext(ctx)
	}
	pair, _, err := s.client.KV().Get(key, opts)
	if err != nil {
		return nil, false, err
	}
	if pair == nil {
		s.updateKey(key, nil)
		return nil, false, nil
	}
	val := make([]byte, len(pair.Value))
	copy(val, pair.Value)
	s.updateKey(key, val)
	return val, true, nil
}

func (s *ConfigStore) watchPrefix(prefix string) {
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

func (s *ConfigStore) pullPrefix(prefix string, lastIndex uint64) (uint64, error) {
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

	var events []ConfigEvent
	s.mu.Lock()
	for k, v := range updates {
		old, ok := s.data[k]
		if !ok || !bytesEqual(old, v) {
			s.data[k] = v
			events = append(events, ConfigEvent{Key: k, Data: v})
		}
	}
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			if _, ok := updates[k]; !ok {
				delete(s.data, k)
				events = append(events, ConfigEvent{Key: k, Data: nil})
			}
		}
	}
	subs, handlers := s.collectNotifyTargetsLocked(events)
	s.mu.Unlock()
	s.notify(subs, handlers, events)

	if meta == nil || meta.LastIndex == 0 {
		return lastIndex, nil
	}
	return meta.LastIndex, nil
}

func (s *ConfigStore) watchKey(key string) {
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

		nextIndex, err := s.pullKey(key, lastIndex)
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

func (s *ConfigStore) pullKey(key string, lastIndex uint64) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 310*time.Second)
	defer cancel()

	opts := (&api.QueryOptions{
		WaitIndex: lastIndex,
		WaitTime:  300 * time.Second,
	}).WithContext(ctx)

	pair, meta, err := s.client.KV().Get(key, opts)
	if err != nil {
		return lastIndex, err
	}

	var val []byte
	if pair != nil {
		val = make([]byte, len(pair.Value))
		copy(val, pair.Value)
	}
	s.updateKey(key, val)

	if meta == nil || meta.LastIndex == 0 {
		return lastIndex, nil
	}
	return meta.LastIndex, nil
}

func (s *ConfigStore) updateKey(key string, value []byte) {
	var events []ConfigEvent
	s.mu.Lock()
	old, ok := s.data[key]
	if value == nil {
		if ok {
			delete(s.data, key)
			events = append(events, ConfigEvent{Key: key, Data: nil})
		}
	} else {
		if !ok || !bytesEqual(old, value) {
			s.data[key] = value
			events = append(events, ConfigEvent{Key: key, Data: value})
		}
	}
	subs, handlers := s.collectNotifyTargetsLocked(events)
	s.mu.Unlock()
	s.notify(subs, handlers, events)
}

func (s *ConfigStore) collectNotifyTargetsLocked(events []ConfigEvent) ([]chan ConfigEvent, map[string][]func(ConfigEvent)) {
	if len(events) == 0 {
		return nil, nil
	}
	subs := make([]chan ConfigEvent, 0, len(s.subs))
	for ch := range s.subs {
		subs = append(subs, ch)
	}
	handlers := make(map[string][]func(ConfigEvent))
	for i := range events {
		ev := events[i]
		if m := s.handlers[ev.Key]; m != nil {
			for _, fn := range m {
				handlers[ev.Key] = append(handlers[ev.Key], fn)
			}
		}
	}
	return subs, handlers
}

func (s *ConfigStore) notify(subs []chan ConfigEvent, handlers map[string][]func(ConfigEvent), events []ConfigEvent) {
	if len(events) == 0 {
		return
	}
	for _, ch := range subs {
		for i := range events {
			select {
			case ch <- events[i]:
			default:
			}
		}
	}
	if len(handlers) == 0 {
		return
	}
	for i := range events {
		ev := events[i]
		if fns := handlers[ev.Key]; len(fns) > 0 {
			for _, fn := range fns {
				func() {
					defer func() { _ = recover() }()
					fn(ev)
				}()
			}
		}
	}
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

func nilWriteOptions(ctx context.Context) *api.WriteOptions {
	if ctx == nil {
		return nil
	}
	return (&api.WriteOptions{}).WithContext(ctx)
}
