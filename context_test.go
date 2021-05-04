// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_Next(t *testing.T) {
	r := newRouter(newContext)

	var buf bytes.Buffer
	r.Get("/",
		func(c Context) {
			buf.WriteString("foo")
			c.Next()
			buf.WriteString("foo2")
		},
		func(c Context) {
			buf.WriteString("bar")
		},
	)

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "foobarfoo2", buf.String())
}
