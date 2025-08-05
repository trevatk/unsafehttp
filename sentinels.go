package unsafehttp

import "errors"

var (
	// ErrUnsupportedHttpVersion
	ErrUnsupportedHttpVersion = errors.New("unsupported http version")
	// ErrRouteNotFound
	ErrRouteNotFound = errors.New("route not found")
	// ErrRequestBodyTooLarge
	ErrRequestBodyTooLarge = errors.New("request body is larger than server limit")
)
