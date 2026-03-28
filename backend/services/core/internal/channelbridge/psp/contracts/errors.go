package contracts

import "errors"

var (
	ErrNoDriver     = errors.New("psp: no driver registered for key")
	ErrUnsupported  = errors.New("psp: operation not supported by this driver")
	ErrVerifyNotify = errors.New("psp: notify verify failed")
)
