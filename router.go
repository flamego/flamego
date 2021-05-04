// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flamego/flamego/internal/route"
)

// Router is the router for adding routes and their handlers.
type Router interface {
	// AutoHead sets a boolean value which determines whether to add HEAD method
	// automatically when GET method is added. Only routes that are added after call
	// of this method will be affected, existing routes remain unchanged.
	AutoHead(v bool)
	// HandlerWrapper sets handlerWrapper for the router. It is used to wrap Handler
	// and inject logic, and is especially useful for wrapping the Handler to
	// inject.FastInvoker.
	HandlerWrapper(f func(Handler) Handler)
	// Route adds the new route path and its handlers to the router tree.
	Route(method, routePath string, handlers []Handler) *Route
	// Combo returns a ComboRoute for adding handlers of different HTTP methods to
	// the same route.
	Combo(routePath string, handlers ...Handler) *ComboRoute
	// Group pushes a new group with the given route path and its handlers, it then
	// pops the group when leaves the scope of `fn`.
	Group(routePath string, fn func(), handlers ...Handler)
	// Get is a shortcut for `r.Route("GET", routePath, handlers)`.
	Get(routePath string, handlers ...Handler) *Route
	// Patch is a shortcut for `r.Route("PATCH", routePath, handlers)`.
	Patch(routePath string, handlers ...Handler) *Route
	// Post is a shortcut for `r.Route("POST", routePath, handlers)`.
	Post(routePath string, handlers ...Handler) *Route
	// Put is a shortcut for `r.Route("PUT", routePath, handlers)`.
	Put(routePath string, handlers ...Handler) *Route
	// Delete is a shortcut for `r.Route("DELETE", routePath, handlers)`.
	Delete(routePath string, handlers ...Handler) *Route
	// Options is a shortcut for `r.Route("OPTIONS", routePath, handlers)`.
	Options(routePath string, handlers ...Handler) *Route
	// Head is a shortcut for `r.Route("HEAD", routePath, handlers)`.
	Head(routePath string, handlers ...Handler) *Route
	// Connect is a shortcut for `r.Route("CONNECT", routePath, handlers)`.
	Connect(routePath string, handlers ...Handler) *Route
	// Trace is a shortcut for `r.Route("TRACE", routePath, handlers)`.
	Trace(routePath string, handlers ...Handler) *Route
	// Any is a shortcut for `r.Route("*", routePath, handlers)`.
	Any(routePath string, handlers ...Handler) *Route
	// Routes is a shortcut of adding same handlers for different HTTP methods.
	//
	// Example:
	//	m.Routes("/", "GET,POST", handlers)
	Routes(routePath, methods string, handlers ...Handler) *Route
	// NotFound configures a http.HandlerFunc to be called when no matching route is
	// found. When it is not set, http.NotFound is used. Be sure to set
	// http.StatusNotFound as the response status code in your last handler.
	NotFound(handlers ...Handler)
	// URLPath builds the "path" portion of URL with given pairs of values. To
	// include the optional segment, pass `"withOptional", "true"`.
	URLPath(name string, pairs ...string) string
	// ServeHTTP implements the method of http.Handler.
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type contextCreator func(http.ResponseWriter, *http.Request, route.Params, []Handler) Context

type router struct {
	parser       *route.Parser                    // The route parser.
	autoHead     bool                             // Whether to automatically attach the same handler of a GET method as HEAD.
	groups       []group                          // The living stack of nested route groups.
	routeTrees   map[string]route.Tree            // A set of route trees, keys are HTTP methods.
	namedRoutes  map[string]route.Leaf            // A set of named routes.
	staticRoutes map[string]map[string]route.Leaf // A set of static routes, keys are HTTP methods and full route paths.

	notFound http.HandlerFunc // The handler to be called when a route has no match.

	// contextCreator is used to create new Context for incoming requests.
	contextCreator contextCreator

	// handlerWrapper is used to wrap Handler and inject logic, and is especially
	// useful for wrapping the Handler to inject.FastInvoker.
	handlerWrapper func(Handler) Handler
}

// See IETF RFC 7231 and RFC 5789 for supported methods
var httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "CONNECT", "TRACE"}

// newRouter creates and returns a new Router.
func newRouter(contextCreator contextCreator) Router {
	parser, err := route.NewParser()
	if err != nil {
		panic("new parser: " + err.Error())
	}

	r := &router{
		parser:         parser,
		routeTrees:     make(map[string]route.Tree),
		namedRoutes:    make(map[string]route.Leaf),
		staticRoutes:   make(map[string]map[string]route.Leaf),
		contextCreator: contextCreator,
	}
	for _, m := range httpMethods {
		r.routeTrees[m] = route.NewTree()
		r.staticRoutes[m] = make(map[string]route.Leaf)
	}

	r.NotFound(http.NotFound)
	return r
}

func (r *router) AutoHead(v bool) {
	r.autoHead = v
}

func (r *router) HandlerWrapper(f func(Handler) Handler) {
	r.handlerWrapper = f
}

// Route is a wrapper of the route leaf and its router.
type Route struct {
	router *router
	leaf   route.Leaf
}

// Name sets the name for the route.
func (r *Route) Name(name string) {
	if name == "" {
		panic("empty route name")
	} else if _, ok := r.router.namedRoutes[name]; ok {
		panic("duplicated route name: " + name)
	}
	r.router.namedRoutes[name] = r.leaf
}

func (r *router) addRoute(method, routePath string, handler route.Handler) *Route {
	method = strings.ToUpper(method)

	var methods []string
	if method == "*" {
		methods = httpMethods
	} else {
		for _, m := range httpMethods {
			if method == m {
				methods = []string{method}
				break
			}
		}
	}
	if len(methods) == 0 {
		panic("unknown HTTP method: " + method)
	}

	ast, err := r.parser.Parse(routePath)
	if err != nil {
		panic(fmt.Sprintf("unable to parse route %q: %v", routePath, err))
	}

	var leaf route.Leaf
	for _, m := range methods {
		leaf, err = route.AddRoute(r.routeTrees[m], ast, handler)
		if err != nil {
			panic(fmt.Sprintf("unable to add route %q: %v", routePath, err))
		}

		if leaf.Static() {
			r.staticRoutes[m][leaf.Route()] = leaf
		}
	}

	return &Route{
		router: r,
		leaf:   leaf,
	}
}

// group contains information of a nested routing group.
type group struct {
	path     string
	handlers []Handler
}

func (r *router) Route(method, routePath string, handlers []Handler) *Route {
	if len(r.groups) > 0 {
		groupPath := ""
		hs := make([]Handler, 0)
		for _, g := range r.groups {
			groupPath += g.path
			hs = append(hs, g.handlers...)
		}

		routePath = groupPath + routePath
		handlers = append(hs, handlers...)
	}

	validateAndWrapHandlers(handlers, r.handlerWrapper)
	return r.addRoute(method, routePath, func(w http.ResponseWriter, req *http.Request, params route.Params) {
		r.contextCreator(w, req, params, handlers).run()
	})
}

func (r *router) Group(routePath string, fn func(), handlers ...Handler) {
	r.groups = append(r.groups,
		group{
			path:     routePath,
			handlers: handlers,
		},
	)
	fn()
	r.groups = r.groups[:len(r.groups)-1]
}

func (r *router) Get(routePath string, handlers ...Handler) *Route {
	route := r.Route("GET", routePath, handlers)
	if r.autoHead {
		r.Head(routePath, handlers...)
	}
	return route
}

func (r *router) Patch(routePath string, handlers ...Handler) *Route {
	return r.Route("PATCH", routePath, handlers)
}

func (r *router) Post(routePath string, handlers ...Handler) *Route {
	return r.Route("POST", routePath, handlers)
}

func (r *router) Put(routePath string, handlers ...Handler) *Route {
	return r.Route("PUT", routePath, handlers)
}

func (r *router) Delete(routePath string, handlers ...Handler) *Route {
	return r.Route("DELETE", routePath, handlers)
}

func (r *router) Options(routePath string, handlers ...Handler) *Route {
	return r.Route("OPTIONS", routePath, handlers)
}

func (r *router) Head(routePath string, handlers ...Handler) *Route {
	return r.Route("HEAD", routePath, handlers)
}

func (r *router) Connect(routePath string, handlers ...Handler) *Route {
	return r.Route("CONNECT", routePath, handlers)
}

func (r *router) Trace(routePath string, handlers ...Handler) *Route {
	return r.Route("TRACE", routePath, handlers)
}

func (r *router) Any(routePath string, handlers ...Handler) *Route {
	return r.Route("*", routePath, handlers)
}

func (r *router) Routes(routePath, methods string, handlers ...Handler) *Route {
	if methods == "" {
		panic("empty methods")
	}

	var route *Route
	for _, m := range strings.Split(methods, ",") {
		route = r.Route(strings.TrimSpace(m), routePath, handlers)
	}
	return route
}

func (r *router) NotFound(handlers ...Handler) {
	validateAndWrapHandlers(handlers, r.handlerWrapper)
	r.notFound = func(w http.ResponseWriter, req *http.Request) {
		r.contextCreator(w, req, nil, handlers).run()
	}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Fast path for static routes
	leaf, ok := r.staticRoutes[req.Method][req.URL.Path]
	if ok {
		leaf.Handler()(w, req, route.Params{
			"route": leaf.Route(),
		})
		return
	}

	leaf, params, ok := r.routeTrees[req.Method].Match(req.URL.EscapedPath())
	if !ok {
		r.notFound(w, req)
		return
	}

	params["route"] = leaf.Route()
	leaf.Handler()(w, req, params)
}

func (r *router) URLPath(name string, pairs ...string) string {
	leaf, ok := r.namedRoutes[name]
	if !ok {
		panic("route with given name does not exist: " + name)
	}

	vals := make(map[string]string, len(pairs)/2)
	for i := 1; i < len(pairs); i += 2 {
		vals[pairs[i-1]] = pairs[i]
	}

	withOptional := false
	if vals["withOptional"] == "true" {
		withOptional = true
		delete(vals, "withOptional")
	}
	return leaf.URLPath(vals, withOptional)
}

// Combo creates and returns new ComboRoute with common handlers for the route.
func (r *router) Combo(routePath string, handlers ...Handler) *ComboRoute {
	return &ComboRoute{
		router:    r,
		routePath: routePath,
		handlers:  handlers,
		added:     make(map[string]struct{}),
	}
}

// ComboRoute is a wrapper of the router for adding handlers of different HTTP
// methods to the same route.
type ComboRoute struct {
	router    *router
	routePath string
	handlers  []Handler

	added     map[string]struct{}
	lastRoute *Route
}

func (r *ComboRoute) route(fn func(string, ...Handler) *Route, method string, handlers ...Handler) *ComboRoute {
	_, ok := r.added[method]
	if ok {
		panic(fmt.Sprintf("duplicated method %q for route %q", method, r.routePath))
	}
	r.added[method] = struct{}{}

	r.lastRoute = fn(r.routePath, append(r.handlers, handlers...)...)
	return r
}

// Get adds handlers of the GET method to the route.
func (r *ComboRoute) Get(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Get, "GET", handlers...)
}

// Patch adds handlers of the PATCH method to the route.
func (r *ComboRoute) Patch(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Patch, "PATCH", handlers...)
}

// Post adds handlers of the POST method to the route.
func (r *ComboRoute) Post(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Post, "POST", handlers...)
}

// Put adds handlers of the PUT method to the route.
func (r *ComboRoute) Put(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Put, "PUT", handlers...)
}

// Delete adds handlers of the DELETE method to the route.
func (r *ComboRoute) Delete(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Delete, "DELETE", handlers...)
}

// Options adds handlers of the OPTIONS method to the route.
func (r *ComboRoute) Options(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Options, "OPTIONS", handlers...)
}

// Head adds handlers of the HEAD method to the route.
func (r *ComboRoute) Head(handlers ...Handler) *ComboRoute {
	return r.route(r.router.Head, "HEAD", handlers...)
}

// Name sets the name for the route.
func (r *ComboRoute) Name(name string) {
	if r.lastRoute == nil {
		panic("no route has been added")
	}
	r.lastRoute.Name(name)
}
