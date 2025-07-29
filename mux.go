package unsafehttp

import (
	"strings"
)

type route struct {
	method   Method
	pattern  string
	segments []string
}

// Mux
type Mux struct {
	routes map[*route]Handler
}

// NewServeMux
func NewServeMux() *Mux {
	return &Mux{
		routes: make(map[*route]Handler),
	}
}

// Chain
func (m *Mux) Chain(interceptors ...Middleware) {
	// build middleware chain
	// iterate over intercepors and create a chain of
	// middleware calls for each route
	for route := range m.routes {
		for _, it := range interceptors {
			// do not range route
			// we need handler func to persist with each iteration
			m.routes[route] = it(m.routes[route])
		}
	}
}

// Get
func (m *Mux) Get(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(&route{method: "GET", pattern: pattern}, handler)
}

// Post
func (m *Mux) Post(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(&route{method: "POST", pattern: pattern}, handler)
}

// Put
func (m *Mux) Put(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(&route{method: "PUT", pattern: pattern}, handler)
}

// Patch
func (m *Mux) Patch(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(&route{method: "PATCH", pattern: pattern}, handler)
}

// Delete
func (m *Mux) Delete(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(&route{method: "DELETE", pattern: pattern}, handler)
}

func (m *Mux) addRoute(r *route, handlerFunc func(ResponseWriter, *Request)) {
	r.pattern = normalizePath(r.pattern)
	r.segments = splitPattern(r.pattern)
	m.routes[r] = &handler{fn: handlerFunc}
}

// iterate over routes and match request path to mux routes
func (m Mux) matchRoute(req *Request) (Handler, bool) {
	segments := splitPattern(req.Path)
OUTER:
	for route, handler := range m.routes {

		if route.method != req.Method {
			continue
		}

		var (
			match  = false
			params = make(map[string]string, 0)
		)

	INNER:
		for i, ps := range route.segments {
			// possible nil pointer
			// verify iterator is within bounds of segments
			if i > len(segments)-1 {
				continue
			}

			s := segments[i]

			// extract param from path
			if strings.HasPrefix(ps, "{") && strings.HasSuffix(ps, "}") {
				// assign param value
				params[ps] = s

				// if last element in slice
				// set match true and break inner loop
				if i == len(segments)-1 {
					match = true
					break INNER
				}
				continue INNER
			}

			// route segment does not match
			if ps != s {
				continue OUTER
			}

			// if last element in slice
			// set match true
			if i == len(segments)-1 {
				match = true
			}
		}

		if match {
			// set request params
			req.Params = params
			// match found
			return handler, true
		}
	}

	// not found
	return nil, false
}

func normalizePath(s string) string {
	return strings.ToValidUTF8(strings.ToLower(strings.TrimSpace(s)), "")
}

func splitPattern(pattern string) []string {
	segments := make([]string, 0)
	// split the pattern by `/`
	// if value prefix `{` and suffix `}` is param
	// else segment
	for s := range strings.SplitSeq(pattern, "/") {
		segments = append(segments, s)
	}
	return segments
}
