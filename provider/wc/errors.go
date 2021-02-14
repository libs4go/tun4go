package wc

import "github.com/libs4go/errors"

// ScopeOfAPIError .
const errVendor = "wc"

// errors
var (
	ErrURLKey       = errors.New("url key not found", errors.WithCode(-1), errors.WithVendor(errVendor))
	ErrURLBridge    = errors.New("url bridge not found", errors.WithCode(-2), errors.WithVendor(errVendor))
	ErrHMAC         = errors.New("hmac compare mismatch", errors.WithCode(-3), errors.WithVendor(errVendor))
	ErrMessage      = errors.New("unexpect message", errors.WithCode(-4), errors.WithVendor(errVendor))
	ErrFormat       = errors.New("message format error", errors.WithCode(-5), errors.WithVendor(errVendor))
	ErrStatus       = errors.New("Tunnel status error", errors.WithCode(-6), errors.WithVendor(errVendor))
	ErrParams       = errors.New("tunnel create params error", errors.WithCode(-7), errors.WithVendor(errVendor))
	ErrDisconnected = errors.New("tunnel peer disconnect", errors.WithCode(-8), errors.WithVendor(errVendor))
)
