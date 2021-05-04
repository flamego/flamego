// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"github.com/flamego/flamego/internal/inject"
	"github.com/flamego/flamego/internal/route"
)

type mockContext struct {
	inject.Injector
	ResponseWriter

	params route.Params

	setParent_ func(inject.Injector)
	urlPath_   urlPather
	written_   func() bool
	next_      func()
	run_       func()
}

func newMockContext() *mockContext {
	return &mockContext{
		Injector: inject.New(),
	}
}

func (c *mockContext) SetParent(injector inject.Injector) {
	c.setParent_(injector)
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

func (c *mockContext) run() {
	c.run_()
}
