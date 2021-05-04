// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/flamego/flamego/internal/inject"
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
			panic(fmt.Sprintf("unable to invoke %dth handler: %v", c.index, err))
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
