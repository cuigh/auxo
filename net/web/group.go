package web

import "net/http"

var _ Router = &Group{}

// Group is a set of sub-routes which is associated with a prefix and shares filters.
type Group struct {
	prefix  string
	filters []Filter
	server  *Server
}

// Group creates a new router group.
func (g *Group) Group(prefix string, filters ...Filter) *Group {
	return g.server.Group(g.prefix+prefix, g.mergeFilters(filters)...)
}

// Use adds filters to the group routes.
func (g *Group) Use(filters ...Filter) {
	g.filters = append(g.filters, filters...)
}

// UseFunc adds filters to the router.
func (g *Group) UseFunc(filters ...FilterFunc) {
	for _, f := range filters {
		g.filters = append(g.filters, f)
	}
}

// Connect registers a route that matches 'CONNECT' method.
func (g *Group) Connect(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodConnect, path, h, filters...)
}

// Delete registers a route that matches 'DELETE' method.
func (g *Group) Delete(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodDelete, path, h, filters...)
}

// Get registers a route that matches 'GET' method.
func (g *Group) Get(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodGet, path, h, filters...)
}

// Head registers a route that matches 'HEAD' method.
func (g *Group) Head(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodHead, path, h, filters...)
}

// Options registers a route that matches 'OPTIONS' method.
func (g *Group) Options(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodOptions, path, h, filters...)
}

// Patch registers a route that matches 'PATCH' method.
func (g *Group) Patch(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodPatch, path, h, filters...)
}

// Post registers a route that matches 'POST' method.
func (g *Group) Post(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodPost, path, h, filters...)
}

// Put registers a route that matches 'PUT' method.
func (g *Group) Put(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodPut, path, h, filters...)
}

// Trace registers a route that matches 'TRACE' method.
func (g *Group) Trace(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.add(http.MethodTrace, path, h, filters...)
}

// Any registers a route that matches all the HTTP methods.
func (g *Group) Any(path string, handler HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.server.Any(g.prefix+path, handler, g.mergeFilters(filters)...)
}

// Match registers a route that matches specific methods.
func (g *Group) Match(methods []string, path string, handler HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.server.Match(methods, g.prefix+path, handler, g.mergeFilters(filters)...)
}

// Handle registers routes from controller.
// It panics if controller's Kind is not Struct.
func (g *Group) Handle(path string, controller interface{}, filters ...Filter) {
	g.server.Handle(g.prefix+path, controller, g.mergeFilters(filters)...)
}

// Static serves files from the given file system root.
func (g *Group) Static(prefix, root string) {
	g.server.Static(prefix, root)
}

// File registers a route in order to server a single file of the local filesystem.
func (g *Group) File(path, file string) {
	g.server.File(g.prefix+path, file)
}

// FileSystem serves files from a custom file system.
func (g *Group) FileSystem(path string, fs http.FileSystem) {
	g.server.FileSystem(g.prefix+path, fs)
}

func (g *Group) add(method, path string, handler HandlerFunc, filters ...Filter) HandlerCustomizer {
	return g.server.add(method, g.prefix+path, handler, g.mergeFilters(filters)...)
}

func (g *Group) mergeFilters(filters []Filter) []Filter {
	var fs []Filter
	fs = append(fs, g.filters...)
	fs = append(fs, filters...)
	return fs
}
