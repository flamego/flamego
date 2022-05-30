// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

var _ internalContext = (*mockContext)(nil)

type mockContext struct {
	*MockContext

	setAction_ func(Handler)
	run_       func()
}

func newMockContext() *mockContext {
	return &mockContext{
		MockContext: NewMockContext(),
	}
}

func (c *mockContext) setAction(h Handler) {
	c.setAction_(h)
}

func (c *mockContext) run() {
	c.run_()
}
