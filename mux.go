package unsafehttp

import "bytes"

type route struct {
	method   []byte
	pattern  []byte
	children map[byte]*route
	isLeaf   bool
	handler  HandlerFunc
}

// Mux
type Mux struct {
	root *route
}

// NewServeMux
func NewServeMux() *Mux {
	return &Mux{
		root: &route{
			children: make(map[byte]*route),
		},
	}
}

// Chain
func (m *Mux) Chain(interceptors ...Middleware) {
	// build middleware chain
	// iterate over intercepors and create a chain of
	// middleware calls for each route
	//for route := range m.routes {
	//	for _, it := range interceptors {
	// do not range route
	// we need handler func to persist with each iteration
	// m.routes[route] = it(m.routes[route])
	//	}
	//}
}

// Get
func (m *Mux) Get(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(m.root, pattern, "GET", handler)
}

// Post
func (m *Mux) Post(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(m.root, pattern, "POST", handler)
}

// Put
func (m *Mux) Put(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(m.root, pattern, "PUT", handler)
}

// Patch
func (m *Mux) Patch(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(m.root, pattern, "PATCH", handler)
}

// Delete
func (m *Mux) Delete(pattern string, handler func(ResponseWriter, *Request)) {
	m.addRoute(m.root, pattern, "DELETE", handler)
}

func (m *Mux) addRoute(node *route, pattern string, method string, handlerFn HandlerFunc) *route {
	if node == nil {
		// root node is nil
		// create root node as leaf
		return &route{
			method:   []byte(method),
			pattern:  []byte(pattern),
			children: make(map[byte]*route),
			isLeaf:   true,
			handler:  handlerFn,
		}
	}

	if pattern == "" {
		// node becomes leaf
		node.isLeaf = true
		node.handler = handlerFn
		return node
	}

	battern := []byte(pattern)

	cpl := commonPrefix(node.pattern, battern)

	if cpl < len(node.pattern) {
		// pattern matches the nodes pattern
		// need to split current node
		return splitRoute(node, battern, method, handlerFn, cpl)
	}

	if cpl == len(node.pattern) && cpl < len(pattern) {
		// pattern included in node pattern
		// continue to next level
		r := pattern[cpl:]
		fb := r[0]
		node.children[fb] = m.addRoute(node.children[fb], r, method, handlerFn)
		return node
	}

	if cpl == len(node.pattern) && cpl == len(pattern) {
		// pattern and node pattern are the same
		// update values
		node.isLeaf = true
		node.handler = handlerFn
		node.method = []byte(method)
		return node
	}

	return node
}

func splitRoute(node *route, pattern []byte, method string, handlerFn HandlerFunc, commonPrefixLength int) *route {
	nn := &route{
		pattern:  node.pattern[:commonPrefixLength],
		children: make(map[byte]*route),
		method:   []byte(method),
	}

	// existing node becomes child of new node
	node.pattern = node.pattern[commonPrefixLength:]
	nn.children[node.pattern[0]] = node

	if commonPrefixLength < len(pattern) {
		// rest of key becomes prefix of new child node
		c := &route{
			pattern:  pattern[commonPrefixLength:],
			isLeaf:   true,
			method:   []byte(method),
			handler:  handlerFn,
			children: make(map[byte]*route),
		}
		nn.children[c.pattern[0]] = c
	} else {
		// pattern matches new node pattern
		nn.isLeaf = true
		nn.handler = handlerFn
	}

	return nn
}

func commonPrefix(a, b []byte) int {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return i
}

// iterate over routes and match request path to mux routes
func (m Mux) matchRoute(req *Request) (HandlerFunc, bool) {
	n := m.root
	pattern := req.Path
	for len(pattern) > 0 {
		if n == nil {
			// root route does not exist
			return nil, false
		}

		cpl := commonPrefix(pattern, n.pattern)
		if cpl == 0 {
			if child, ok := n.children[pattern[0]]; ok {
				// path compression can cause miss
				// check children
				n = child
				continue
			}
			// there are no common patterns
			return nil, false
		}

		if cpl == len(n.pattern) && cpl == len(pattern) {
			if bytes.Equal(n.method, req.Method) {
				// match found
				return n.handler, true
			}
		}

		if cpl == len(n.pattern) && cpl < len(pattern) {
			// if pattern matches node
			// continue to next level
			pattern = pattern[cpl:]
			fc := pattern[0]
			if child, ok := n.children[fc]; ok {
				n = child
				continue
			}
			return nil, false
		}
	}

	if n != nil && n.isLeaf {
		return n.handler, true
	}

	return nil, false
}
