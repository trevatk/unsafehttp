package unsafehttp

// Handler
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP
func (fn HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	fn(w, r)
}
