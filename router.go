package unsafehttp

import "bytes"

// WalkFunc
type WalkFunc func(string, string, HandlerFunc)

type route struct {
	method   []byte
	pattern  []byte
	children map[byte]*route
	isLeaf   bool
	handler  HandlerFunc
}

// Router
type Router interface {
	// Get
	Get(string, HandlerFunc)
	// Post
	Post(string, HandlerFunc)
	// Put
	Put(string, HandlerFunc)
	// Patch
	Patch(string, HandlerFunc)
	// Delete
	Delete(string, HandlerFunc)

	// Group
	Group(string, GroupFunc)

	// Chain
	Use(...Middleware)

	Walk(WalkFunc)

	// package level

	matchRoute(*Request) (HandlerFunc, bool)
}

// Router
type router struct {
	root   *route
	prefix string
	mws    []Middleware
}

// NewRouter
func NewRouter() Router {
	return &router{
		root: &route{
			children: make(map[byte]*route),
		},
		prefix: "",
		mws:    make([]Middleware, 0),
	}
}

// Use
func (r *router) Use(mws ...Middleware) {
	r.mws = append(r.mws, mws...)
}

// Get
func (r *router) Get(pattern string, handler HandlerFunc) {
	handler = chain(handler, r.mws)
	r.addRoute(r.root, r.prefix+pattern, "GET", handler)
}

// Post
func (r *router) Post(pattern string, handler HandlerFunc) {
	handler = chain(handler, r.mws)
	r.addRoute(r.root, r.prefix+pattern, "POST", handler)
}

// Put
func (r *router) Put(pattern string, handler HandlerFunc) {
	handler = chain(handler, r.mws)
	r.addRoute(r.root, r.prefix+pattern, "PUT", handler)
}

// Patch
func (r *router) Patch(pattern string, handler HandlerFunc) {
	handler = chain(handler, r.mws)
	r.addRoute(r.root, r.prefix+pattern, "PATCH", handler)
}

// Delete
func (r *router) Delete(pattern string, handler HandlerFunc) {
	handler = chain(handler, r.mws)
	r.addRoute(r.root, r.prefix+pattern, "DELETE", handler)
}

func (r *router) addRoute(node *route, pattern string, method string, handlerFn HandlerFunc) *route {
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
		p := pattern[cpl:]
		node.children[p[0]] = r.addRoute(node.children[p[0]], p, method, handlerFn)
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

// iterate over routes and match request path to mux routes
func (r *router) matchRoute(req *Request) (HandlerFunc, bool) {
	n := r.root
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

// Walk
func (r *router) Walk(fn WalkFunc) {
	r.root.walk("", fn)
}

func (ro *route) walk(path string, fn WalkFunc) {
	if ro == nil {
		return
	}

	p := path + string(ro.pattern)
	if ro.isLeaf {
		fn(string(ro.method), p, ro.handler)
	}

	for _, child := range ro.children {
		child.walk(p, fn)
	}
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

func chain(handler HandlerFunc, mws []Middleware) HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}
