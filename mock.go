// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"

	"github.com/flamego/flamego/internal/route"
)

type mockContext struct {
	handlers []Handler

	responseWriter http.ResponseWriter
	request        *http.Request
	params         route.Params

	run_ func()
}

func (c *mockContext) run() {
	if c.run_ != nil {
		c.run_()
	}
}
