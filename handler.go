package unsafehttp

// Handler
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type handler struct {
	fn func(ResponseWriter, *Request)
}

// ServeHTTP
func (h *handler) ServeHTTP(w ResponseWriter, r *Request) {
	h.fn(w, r)
}
