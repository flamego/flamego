// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/flamego/flamego/inject"
	"github.com/flamego/flamego/internal/route"
)

// Context is the runtime context of the coming request, and provide handy
// methods to enhance developer experience.
type Context interface {
	inject.Injector
	// ResponseWriter returns the ResponseWriter in current context.
	ResponseWriter() ResponseWriter
	// Request returns the Request in current context.
	Request() *Request

	// URLPath builds the "path" portion of URL with given pairs of values. To
	// include the optional segment, pass `"withOptional", "true"`.
	//
	// This is a transparent wrapper of Router.URLPath.
	URLPath(name string, pairs ...string) string

	// Next runs the next handler in the context chain.
	Next()
	// RemoteAddr extracts and returns the remote IP address from following attempts
	// in sequence:
	//  - "X-Real-IP" request header
	//  - "X-Forwarded-For" request header
	//  - http.Request.RemoteAddr field
	RemoteAddr() string
	// Redirect sends a redirection to the response to the given location. If the
	// `status` is not given, the http.StatusFound is used.
	Redirect(location string, status ...int)

	// Params returns value of given bind parameter.
	Params(name string) string
	// ParamsInt returns value of given bind parameter parsed as int.
	ParamsInt(name string) int

	// Query queries URL parameter with given name.
	Query(name string) string
	// QueryTrim queries and trims spaces from the value.
	QueryTrim(name string) string
	// QueryStrings returns a list of results with given name.
	QueryStrings(name string) []string
	// QueryUnescape returns unescaped query result.
	QueryUnescape(name string) string
	// QueryBool returns query result in bool type.
	QueryBool(name string) bool
	// QueryInt returns query result in int type.
	QueryInt(name string) int
	// QueryInt64 returns query result in int64 type.
	QueryInt64(name string) int64
	// QueryFloat64 returns query result in float64 type.
	QueryFloat64(name string) float64

	// SetCookie escapes the cookie value and sets it to the current response.
	SetCookie(cookie http.Cookie)
	// Cookie returns the named cookie in the request or empty if not found. If
	// multiple cookies match the given name, only one cookie will be returned. The
	// returned value is unescaped using `url.QueryUnescape`, original value is
	// returned instead if unable to unescape.
	Cookie(name string) string

	// setAction sets the final handler in the context chain.
	setAction(Handler)
	// run executes all handlers in the context chain.
	run()
}

type context struct {
	inject.Injector

	handlers []Handler // The list of handlers to be executed.
	action   Handler   // The last action handler to be executed.
	index    int       // The index of the current handler that is being executed.

	responseWriter ResponseWriter // The http.ResponseWriter wrapper for the coming request.
	request        *Request       // The http.Request wrapper for the coming request.
	params         route.Params   // The values of bind parameters for the coming request.

	// urlPath is used to build URL path for a route.
	urlPath urlPather
}

type urlPather func(name string, pairs ...string) string

// newContext creates and returns a new Context.
func newContext(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
	c := &context{
		Injector:       inject.New(),
		handlers:       handlers,
		responseWriter: NewResponseWriter(r.Method, w),
		request:        &Request{Request: r},
		params:         params,
		urlPath:        urlPath,
	}
	c.Map(c)
	c.MapTo(c.responseWriter, (*http.ResponseWriter)(nil))
	c.Map(r)
	return c
}

func (c *context) ResponseWriter() ResponseWriter {
	return c.responseWriter
}

func (c *context) Request() *Request {
	return c.request
}

func (c *context) URLPath(name string, pairs ...string) string {
	return c.urlPath(name, pairs...)
}

func (c *context) Next() {
	c.index++
	c.run()
}

func (c *context) setAction(h Handler) {
	c.action = h
}

func (c *context) run() {
	for c.index <= len(c.handlers) {
		var h Handler
		if c.index == len(c.handlers) {
			h = c.action
		} else {
			h = c.handlers[c.index]
		}

		if h == nil {
			c.index++
			return
		}

		vals, err := c.Invoke(h)
		if err != nil {
			panic(fmt.Sprintf("unable to invoke %dth handler [%T]: %v", c.index, h, err))
		}
		c.index++

		// If the handler returned something, write it to the response.
		if len(vals) > 0 {
			ev := c.Value(reflect.TypeOf(ReturnHandler(nil)))
			handleReturn := ev.Interface().(ReturnHandler)
			handleReturn(c, vals)
		}

		if c.ResponseWriter().Written() {
			return
		}
	}
}

func (c *context) RemoteAddr() string {
	addr := c.Request().Header.Get("X-Real-IP")
	if addr != "" {
		return addr
	}

	addr = c.Request().Header.Get("X-Forwarded-For")
	if addr != "" {
		return addr
	}

	addr = c.Request().RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > -1 {
		addr = addr[:i]
	}
	return addr
}

func (c *context) Redirect(location string, status ...int) {
	code := http.StatusFound
	if len(status) == 1 {
		code = status[0]
	}

	http.Redirect(c.ResponseWriter(), c.Request().Request, location, code)
}

func (c *context) Params(name string) string {
	return c.params[name]
}

func (c *context) ParamsInt(name string) int {
	i, _ := strconv.Atoi(c.Params(name))
	return i
}

func (c *context) Query(name string) string {
	return c.Request().URL.Query().Get(name)
}

func (c *context) QueryTrim(name string) string {
	return strings.TrimSpace(c.Query(name))
}

func (c *context) QueryStrings(name string) []string {
	for k, v := range c.Request().URL.Query() {
		if k == name {
			return v
		}
	}
	return []string{}
}

func (c *context) QueryUnescape(name string) string {
	v, _ := url.QueryUnescape(c.Query(name))
	return v
}

func (c *context) QueryBool(name string) bool {
	v, _ := strconv.ParseBool(c.Query(name))
	return v
}

func (c *context) QueryInt(name string) int {
	v, _ := strconv.ParseInt(c.Query(name), 10, 0)
	return int(v)
}

func (c *context) QueryInt64(name string) int64 {
	v, _ := strconv.ParseInt(c.Query(name), 10, 64)
	return v
}

func (c *context) QueryFloat64(name string) float64 {
	v, _ := strconv.ParseFloat(c.Query(name), 64)
	return v
}

func (c *context) SetCookie(cookie http.Cookie) {
	cookie.Value = url.QueryEscape(cookie.Value)
	c.ResponseWriter().Header().Add("Set-Cookie", cookie.String())
}

func (c *context) Cookie(name string) string {
	cookie, err := c.Request().Cookie(name)
	if err != nil {
		return ""
	}

	val, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return cookie.Value
	}
	return val
}
