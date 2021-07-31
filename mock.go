// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"strconv"

	"github.com/flamego/flamego/internal/inject"
	"github.com/flamego/flamego/internal/route"
)

var _ Context = (*mockContext)(nil)

type mockContext struct {
	inject.Injector
	responseWriter ResponseWriter
	request        *Request

	params route.Params

	urlPath_    urlPather
	written_    func() bool
	next_       func()
	setAction_  func(Handler)
	run_        func()
	remoteAddr_ func() string
	setCookie_  func(cookie http.Cookie)
	cookie_     func(string) string
}

func newMockContext() *mockContext {
	return &mockContext{
		Injector: inject.New(),
	}
}

func (c *mockContext) ResponseWriter() ResponseWriter {
	return c.responseWriter
}

func (c *mockContext) Request() *Request {
	return c.request
}

func (c *mockContext) URLPath(name string, pairs ...string) string {
	return c.urlPath_(name, pairs...)
}

func (c *mockContext) Written() bool {
	return c.written_()
}

func (c *mockContext) Next() {
	c.next_()
}

func (c *mockContext) setAction(h Handler) {
	c.setAction_(h)
}

func (c *mockContext) run() {
	c.run_()
}

func (c *mockContext) RemoteAddr() string {
	return c.remoteAddr_()
}

func (c *mockContext) Params(name string) string {
	return c.params[name]
}

func (c *mockContext) ParamsInt(name string) int {
	i, _ := strconv.Atoi(c.Params(name))
	return i
}

func (c *mockContext) SetCookie(cookie http.Cookie) {
	c.setCookie_(cookie)
}

func (c *mockContext) Cookie(name string) string {
	return c.cookie_(name)
}
