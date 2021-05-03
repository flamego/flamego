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

// Router is the implementation of the routing middleware.
type Router struct {
	parser       *route.Parser         // The route parser.
	autoHead     bool                  // Whether to automatically attach the same handler of a GET method as HEAD.
	groups       []group               // The living stack of nested route groups.
	routeTrees   map[string]route.Tree // A set of route trees, keys are HTTP methods.
	namedRoutes  map[string]route.Leaf // A set of named routes.
	staticRoutes map[string]route.Leaf // A set of static routes, keys are full route paths.

	notFound http.HandlerFunc // The handler to be called when a route has no match.

	// createContext is used to create new Context for coming requests.
	createContext func(http.ResponseWriter, *http.Request, route.Params, []Handler) Context

	// handlerWrapper is used to wrap Handler and inject logic, and is especially
	// useful for wrapping the Handler to inject.FastInvoker.
	handlerWrapper func(Handler) Handler
}

var httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

// newRouter creates and returns a new Router.
func newRouter() *Router {
	parser, err := route.NewParser()
	if err != nil {
		panic("new parser: " + err.Error())
	}

	r := &Router{
		parser:       parser,
		routeTrees:   make(map[string]route.Tree),
		namedRoutes:  make(map[string]route.Leaf),
		staticRoutes: make(map[string]route.Leaf),
	}
	for _, m := range httpMethods {
		r.routeTrees[m] = route.NewTree()
	}
	return r
}

// AutoHead sets a boolean value which determines whether to add HEAD method
// automatically when GET method is added. Only routes that are added after call
// of this method will be affected, existing routes remain unchanged.
func (r *Router) AutoHead(v bool) {
	r.autoHead = v
}

// HandlerWrapper sets handlerWrapper for the router. It is used to wrap Handler
// and inject logic, and is especially useful for wrapping the Handler to
// inject.FastInvoker.
func (r *Router) HandlerWrapper(f func(Handler) Handler) {
	r.handlerWrapper = f
}

// Route is a wrapper of the route leaf and its router.
type Route struct {
	router *Router
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

func (r *Router) addRoute(method, routePath string, handler route.Handler) *Route {
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
	for _, m := range methods[1:] {
		leaf, err = route.AddRoute(r.routeTrees[m], ast, handler)
		if err != nil {
			panic(fmt.Sprintf("unable to add route %q: %v", routePath, err))
		}
	}

	if leaf.Static() {
		r.staticRoutes[leaf.Route()] = leaf
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

// Route adds the new route path and its handlers to the router tree.
func (r *Router) Route(method, routePath string, handlers []Handler) *Route {
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
		r.createContext(w, req, params, handlers).run()
	})
}

// Group pushes a new group with the given route path and its handlers, it then
// pops the group when leaves the scope of `fn`.
func (r *Router) Group(routePath string, fn func(), handlers ...Handler) {
	r.groups = append(r.groups,
		group{
			path:     routePath,
			handlers: handlers,
		},
	)
	fn()
	r.groups = r.groups[:len(r.groups)-1]
}

// Get is a shortcut for `r.Route("GET", routePath, handlers)`.
func (r *Router) Get(routePath string, handlers ...Handler) *Route {
	route := r.Route("GET", routePath, handlers)
	if r.autoHead {
		r.Head(routePath, handlers)
	}
	return route
}

// Patch is a shortcut for `r.Route("PATCH", routePath, handlers)`.
func (r *Router) Patch(routePath string, handlers ...Handler) *Route {
	return r.Route("PATCH", routePath, handlers)
}

// Post is a shortcut for `r.Route("POST", routePath, handlers)`.
func (r *Router) Post(routePath string, handlers ...Handler) *Route {
	return r.Route("POST", routePath, handlers)
}

// Put is a shortcut for `r.Route("PUT", routePath, handlers)`.
func (r *Router) Put(routePath string, handlers ...Handler) *Route {
	return r.Route("PUT", routePath, handlers)
}

// Delete is a shortcut for `r.Route("DELETE", routePath, handlers)`.
func (r *Router) Delete(routePath string, handlers ...Handler) *Route {
	return r.Route("DELETE", routePath, handlers)
}

// Options is a shortcut for `r.Route("OPTIONS", routePath, handlers)`.
func (r *Router) Options(routePath string, handlers ...Handler) *Route {
	return r.Route("OPTIONS", routePath, handlers)
}

// Head is a shortcut for `r.Route("HEAD", routePath, handlers)`.
func (r *Router) Head(routePath string, handlers ...Handler) *Route {
	return r.Route("HEAD", routePath, handlers)
}

// Any is a shortcut for `r.Route("*", routePath, handlers)`.
func (r *Router) Any(routePath string, handlers ...Handler) *Route {
	return r.Route("*", routePath, handlers)
}

// Routes is a shortcut of adding same handlers for different HTTP methods.
//
// Example:
//	m.Routes("/", "GET,POST", handlers)
func (r *Router) Routes(routePath, methods string, handlers ...Handler) *Route {
	if methods == "" {
		panic("empty methods")
	}

	var route *Route
	for _, m := range strings.Split(methods, ",") {
		route = r.Route(strings.TrimSpace(m), routePath, handlers)
	}
	return route
}

// NotFound configures a http.HandlerFunc to be called when no matching route is
// found. When it is not set, http.NotFound is used. Be sure to set
// http.StatusNotFound as the response status code in your last handler.
func (r *Router) NotFound(handlers ...Handler) {
	validateAndWrapHandlers(handlers, r.handlerWrapper)
	r.notFound = func(w http.ResponseWriter, req *http.Request) {
		r.createContext(w, req, nil, handlers).run()
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Fast path for static routes
	leaf, ok := r.staticRoutes[req.URL.Path]
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

// URLPath builds the "path" portion of URL with given pairs of values. To
// include the optional segment, pass `"withOptional", "true"`.
func (r *Router) URLPath(name string, pairs ...string) string {
	leaf, ok := r.namedRoutes[name]
	if !ok {
		panic("route with given name does not exist: " + name)
	}

	vals := make(map[string]string, len(pairs)/2)
	for i := 1; i <= len(pairs); i += 2 {
		vals[pairs[i]] = pairs[i+1]
	}

	withOptional := false
	if vals["withOptional"] == "true" {
		withOptional = true
		delete(vals, "withOptional")
	}
	return leaf.URLPath(vals, withOptional)
}
