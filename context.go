// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/flamego/flamego/inject"
	"github.com/flamego/flamego/internal/route"
)

// Context is the runtime context of the coming request, and provide handy
// methods to enhance developer experience.
//
//go:generate go-mockgen -f github.com/flamego/flamego -i Context -o mock_context_test.go
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

	// Params returns all bind parameters.
	Params() Params
	// Param returns value of the given bind parameter.
	Param(name string) string
	// ParamInt returns value of the given bind parameter parsed as int.
	ParamInt(name string) int
	// ParamInt64 returns value of the given bind parameter parsed as int64.
	ParamInt64(name string) int64

	// Query returns value of the given URL parameter. The `defaultVal` (when set)
	// or zero value is returned when the given parameter is absent.
	Query(name string, defaultVal ...string) string
	// QueryTrim trims spaces and returns value of the given URL parameter. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryTrim(name string, defaultVal ...string) string
	// QueryStrings returns a list of strings of the given URL parameter. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryStrings(name string, defaultVal ...[]string) []string
	// QueryUnescape returns unescaped query result of the given URL parameter. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryUnescape(name string, defaultVal ...string) string
	// QueryBool returns value of the given URL parameter parsed as bool. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryBool(name string, defaultVal ...bool) bool
	// QueryInt returns value of the given URL parameter parsed as int. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryInt(name string, defaultVal ...int) int
	// QueryInt64 returns value of the given URL parameter parsed as int64. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryInt64(name string, defaultVal ...int64) int64
	// QueryFloat64 returns value of the given URL parameter parsed as float64. The
	// `defaultVal` (when set) or zero value is returned when the given parameter is
	// absent.
	QueryFloat64(name string, defaultVal ...float64) float64

	// SetCookie escapes the cookie value and sets it to the current response.
	SetCookie(cookie http.Cookie)
	// Cookie returns the named cookie in the request or empty if not found. If
	// multiple cookies match the given name, only one cookie will be returned. The
	// returned value is unescaped using `url.QueryUnescape`, original value is
	// returned instead if unable to unescape.
	Cookie(name string) string
}

// internalContext is the wrapper of the Context with private methods.
type internalContext interface {
	Context

	// setAction sets the final handler in the context chain.
	setAction(Handler)
	// run executes all handlers in the context chain.
	run()
}

// Params is a set of bind parameters with their values that are extracted from
// the request path.
type Params map[string]string

type context struct {
	inject.Injector

	handlers []Handler // The list of handlers to be executed.
	action   Handler   // The last action handler to be executed.
	index    int       // The index of the current handler that is being executed.

	responseWriter ResponseWriter // The http.ResponseWriter wrapper for the coming request.
	request        *Request       // The http.Request wrapper for the coming request.
	params         Params         // The values of bind parameters for the coming request.

	// urlPath is used to build URL path for a route.
	urlPath urlPather
}

type urlPather func(name string, pairs ...string) string

// newContext creates and returns a new Context.
func newContext(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) internalContext {
	c := &context{
		Injector:       inject.New(),
		handlers:       handlers,
		responseWriter: NewResponseWriter(r.Method, w),
		request:        &Request{Request: r},
		params:         Params(params),
		urlPath:        urlPath,
	}
	c.MapTo(c, (*Context)(nil))
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

// ordinalize ordinalizes the number by adding the ordinal to the number.
func ordinalize(number int) string {
	abs := int(math.Abs(float64(number)))

	nstr := strconv.Itoa(number)
	i := abs % 100
	if i == 11 || i == 12 || i == 13 {
		return nstr + "th"
	}

	switch abs % 10 {
	case 1:
		return nstr + "st"
	case 2:
		return nstr + "nd"
	case 3:
		return nstr + "rd"
	default:
		return nstr + "th"
	}
}

func (c *context) run() {
	for c.index <= len(c.handlers) {
		// Break out when the request context has been cancelled.
		select {
		case <-c.Request().Context().Done():
			return
		default:
		}

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
			panic(fmt.Sprintf("unable to invoke the %s handler [%s:%T]: %v",
				ordinalize(c.index), runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(), h, err))
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

func (c *context) Params() Params {
	return c.params
}

func (c *context) Param(name string) string {
	return c.params[name]
}

func (c *context) ParamInt(name string) int {
	i, _ := strconv.Atoi(c.Param(name))
	return i
}

func (c *context) ParamInt64(name string) int64 {
	v, _ := strconv.ParseInt(c.Param(name), 10, 64)
	return v
}

func (c *context) Query(name string, defaultVal ...string) string {
	v := c.Request().URL.Query().Get(name)
	if v == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return v
}

func (c *context) QueryTrim(name string, defaultVal ...string) string {
	return strings.TrimSpace(c.Query(name, defaultVal...))
}

func (c *context) QueryStrings(name string, defaultVal ...[]string) []string {
	for k, v := range c.Request().URL.Query() {
		if k == name {
			return v
		}
	}

	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return []string{}
}

func (c *context) QueryUnescape(name string, defaultVal ...string) string {
	v, _ := url.QueryUnescape(c.Query(name, defaultVal...))
	return v
}

func (c *context) QueryBool(name string, defaultVal ...bool) bool {
	v := c.Query(name)
	if v == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	b, _ := strconv.ParseBool(v)
	return b
}

func (c *context) QueryInt(name string, defaultVal ...int) int {
	v := c.Query(name)
	if v == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}

	i, _ := strconv.ParseInt(v, 10, 0)
	return int(i)
}

func (c *context) QueryInt64(name string, defaultVal ...int64) int64 {
	v := c.Query(name)
	if v == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}

	i, _ := strconv.ParseInt(v, 10, 64)
	return i
}

func (c *context) QueryFloat64(name string, defaultVal ...float64) float64 {
	v := c.Query(name)
	if v == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}

	f, _ := strconv.ParseFloat(v, 64)
	return f
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
