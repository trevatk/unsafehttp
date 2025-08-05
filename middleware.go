package unsafehttp

// Middleware
type Middleware func(HandlerFunc) HandlerFunc
