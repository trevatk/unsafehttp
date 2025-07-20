package unsafehttp

import (
	"strings"
)

type route struct {
	method   string
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

// Get
func (m *Mux) Get(pattern string, handler func(ResponseWriter, *Request)) {

	// normalize pattern
	pattern = normalizePath(pattern)
	// create route object with
	// method
	// pattern
	r := &route{
		method:  "GET",
		pattern: pattern,
	}

	r.segments = splitPattern(pattern)
	m.routes[r] = Handler(handler)
}

// Post
func (m *Mux) Post(pattern string, handler func(ResponseWriter, *Request)) {
	// normalize pattern
	pattern = normalizePath(pattern)
	// create route object with
	// method
	// pattern
	r := &route{
		method:  "POST",
		pattern: pattern,
	}

	r.segments = splitPattern(pattern)
	m.routes[r] = Handler(handler)
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
