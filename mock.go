// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"

	"github.com/flamego/flamego/inject"
)

var _ Context = (*mockContext)(nil)

type mockContext struct {
	inject.Injector
	responseWriter_ func() ResponseWriter
	request_        func() *Request

	params Params

	urlPath_ urlPather
	written_ func() bool

	next_       func()
	remoteAddr_ func() string
	redirect_   func(string, ...int)

	params_     func() Params
	param_      func(string) string
	paramInt_   func(string) int
	paramInt64_ func(string) int64

	query_         func(string) string
	queryTrim_     func(string) string
	queryStrings_  func(string) []string
	queryUnescape_ func(string) string
	queryBool_     func(string) bool
	queryInt_      func(string) int
	queryInt64_    func(string) int64
	queryFloat64_  func(string) float64

	setCookie_ func(cookie http.Cookie)
	cookie_    func(string) string

	setAction_ func(Handler)
	run_       func()
}

func newMockContext() *mockContext {
	return &mockContext{
		Injector: inject.New(),
	}
}

func (c *mockContext) ResponseWriter() ResponseWriter {
	return c.responseWriter_()
}

func (c *mockContext) Request() *Request {
	return c.request_()
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

func (c *mockContext) Redirect(location string, status ...int) {
	c.redirect_(location, status...)
}

func (c *mockContext) Params() Params {
	return c.params_()
}

func (c *mockContext) Param(name string) string {
	return c.param_(name)
}

func (c *mockContext) ParamInt(name string) int {
	return c.paramInt_(name)
}

func (c *mockContext) ParamInt64(name string) int64 {
	return c.paramInt64_(name)
}

func (c *mockContext) Query(name string) string {
	return c.query_(name)
}

func (c *mockContext) QueryTrim(name string) string {
	return c.queryTrim_(name)
}

func (c *mockContext) QueryStrings(name string) []string {
	return c.queryStrings_(name)
}

func (c *mockContext) QueryUnescape(name string) string {
	return c.queryUnescape_(name)
}

func (c *mockContext) QueryBool(name string) bool {
	return c.queryBool_(name)
}

func (c *mockContext) QueryInt(name string) int {
	return c.queryInt_(name)
}

func (c *mockContext) QueryInt64(name string) int64 {
	return c.queryInt64_(name)
}

func (c *mockContext) QueryFloat64(name string) float64 {
	return c.queryFloat64_(name)
}

func (c *mockContext) SetCookie(cookie http.Cookie) {
	c.setCookie_(cookie)
}

func (c *mockContext) Cookie(name string) string {
	return c.cookie_(name)
}
