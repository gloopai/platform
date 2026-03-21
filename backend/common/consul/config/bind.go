package config

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func BindJSON[T any](ctx context.Context, s *Store, key string, dst *Var[T]) (func(), error) {
	if s == nil {
		return func() {}, errors.New("store required")
	}
	if dst == nil {
		return func() {}, errors.New("dst required")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return func() {}, errors.New("key required")
	}

	apply := func(data []byte) {
		if data == nil {
			dst.Reset()
			return
		}
		var v T
		if err := json.Unmarshal(data, &v); err != nil {
			return
		}
		dst.Store(v)
	}

	s.WatchKey(key)
	if b, ok := s.GetBytes(key); ok {
		apply(b)
	} else {
		if b, ok, _ := s.Fetch(ctx, key); ok {
			apply(b)
		}
	}

	cancel := s.OnKey(key, func(ev Event) {
		apply(ev.Data)
	})
	return cancel, nil
}
