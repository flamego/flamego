// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"fmt"
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

func TestContext_RemoteAddr(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return c.RemoteAddr()
	})

	tests := []struct {
		name           string
		newRequest     func() *http.Request
		wantRemoteAddr string
	}{
		{
			name: "from request field",
			newRequest: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				assert.Nil(t, err)

				req.RemoteAddr = "127.0.0.1:2830"
				return req
			},
			wantRemoteAddr: "127.0.0.1",
		},
		{
			name: "from X-Forwarded-For",
			newRequest: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				assert.Nil(t, err)

				req.RemoteAddr = "127.0.0.1:2830"
				req.Header.Set("X-Forwarded-For", "192.168.0.1")
				return req
			},
			wantRemoteAddr: "192.168.0.1",
		},
		{
			name: "from X-Real-IP",
			newRequest: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				assert.Nil(t, err)

				req.RemoteAddr = "127.0.0.1:2830"
				req.Header.Set("X-Forwarded-For", "192.168.0.1")
				req.Header.Set("X-Real-IP", "10.0.0.1")
				return req
			},
			wantRemoteAddr: "10.0.0.1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req := test.newRequest()

			f.ServeHTTP(resp, req)

			assert.Equal(t, test.wantRemoteAddr, resp.Body.String())
		})
	}
}

func TestContext_Params(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	tests := []struct {
		route    string
		url      string
		handler  Handler
		wantBody string
	}{
		{
			route: "/params/{string}/{int}",
			url:   "/params/hello/123",
			handler: func(c Context) string {
				return fmt.Sprintf("%s %s", c.Params("string"), c.Params("int"))
			},
			wantBody: "hello 123",
		},
		{
			route: "/params-int/{string}/{int}",
			url:   "/params-int/hello/123",
			handler: func(c Context) string {
				return fmt.Sprintf("%d %d", c.ParamsInt("string"), c.ParamsInt("int"))
			},
			wantBody: "0 123",
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			f.Get(test.route, test.handler)

			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)

			assert.Equal(t, test.wantBody, resp.Body.String())
		})
	}
}
