package unsafehttp

// Group
type GroupFunc func(r Router)

type group struct {
	router *router
	prefix string
	mws    []Middleware
}

// Group
func (r *router) Group(pattern string, fn GroupFunc) {
	g := &group{
		router: r,
		prefix: r.prefix + pattern,
		mws:    make([]Middleware, 0),
	}

	fn(g)
}

// Get
func (g *group) Get(pattern string, handler HandlerFunc) {
	handler = chain(handler, g.mws)
	handler = chain(handler, g.router.mws)
	g.router.addRoute(g.router.root, g.prefix+pattern, "GET", handler)
}

// Post
func (g *group) Post(pattern string, handler HandlerFunc) {
	handler = chain(handler, g.mws)
	handler = chain(handler, g.router.mws)
	g.router.addRoute(g.router.root, g.prefix+pattern, "POST", handler)
}

// Put
func (g *group) Put(pattern string, handler HandlerFunc) {
	handler = chain(handler, g.mws)
	handler = chain(handler, g.router.mws)
	g.router.addRoute(g.router.root, g.prefix+pattern, "PUT", handler)
}

// Patch
func (g *group) Patch(pattern string, handler HandlerFunc) {
	handler = chain(handler, g.mws)
	handler = chain(handler, g.router.mws)
	g.router.addRoute(g.router.root, g.prefix+pattern, "PATCH", handler)
}

// Delete
func (g *group) Delete(pattern string, handler HandlerFunc) {
	handler = chain(handler, g.mws)
	handler = chain(handler, g.router.mws)
	g.router.addRoute(g.router.root, g.prefix+pattern, "DELETE", handler)
}

// Group
func (g *group) Group(pattern string, fn GroupFunc) {
	gg := &group{
		router: g.router,
		prefix: g.prefix + pattern,
		mws:    make([]Middleware, 0),
	}

	fn(gg)
}

// Chain
func (g *group) Use(mws ...Middleware) {
	g.router.mws = append(g.router.mws, mws...)
}

// Walk not implemented
func (g *group) Walk(fn WalkFunc) {
	g.router.root.walk(g.prefix, fn)
}

// interface compliance
func (g *group) matchRoute(*Request) (HandlerFunc, bool) {
	return nil, false
}
