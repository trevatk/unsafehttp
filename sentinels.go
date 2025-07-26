package unsafehttp

import "errors"

var (
	// ErrUnsupportedHttpVersion
	ErrUnsupportedHttpVersion = errors.New("unsupported http version")
	// ErrRouteNotFound
	ErrRouteNotFound = errors.New("route not found")
)
