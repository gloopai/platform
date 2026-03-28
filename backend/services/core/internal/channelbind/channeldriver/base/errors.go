package base

import "errors"

var (
	// ErrNoDriver is returned when Registry has no implementation for a driver_key.
	ErrNoDriver = errors.New("channeldriver: no driver registered for key")

	// ErrUnsupported is returned when a driver does not implement an optional operation.
	ErrUnsupported = errors.New("channeldriver: operation not supported by this driver")

	// ErrVerifyNotify is returned when callback signature or payload validation fails.
	ErrVerifyNotify = errors.New("channeldriver: notify verify failed")
)
