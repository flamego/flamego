// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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
	req, err := http.NewRequest(http.MethodGet, "/", nil)
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
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				assert.Nil(t, err)

				req.RemoteAddr = "127.0.0.1:2830"
				return req
			},
			wantRemoteAddr: "127.0.0.1",
		},
		{
			name: "from X-Forwarded-For",
			newRequest: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
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
				req, err := http.NewRequest(http.MethodGet, "/", nil)
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
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)

			assert.Equal(t, test.wantBody, resp.Body.String())
		})
	}
}

func TestContext_Query(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return c.Query("fgq")
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=Flamego&language=Go",
			want: "Flamego",
		},
		{
			name: "empty value",
			url:  "/?fgq=&language=Go",
			want: "",
		},
		{
			name: "multiple value",
			url:  "/?fgq=value1&fgq=value2",
			want: "value1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryTrim(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return c.QueryTrim("fgq")
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=  Flamego  &language=Go",
			want: "Flamego",
		},
		{
			name: "empty value",
			url:  "/?fgq=&language=Go",
			want: "",
		},
		{
			name: "multiple value",
			url:  "/?fgq=  value1&fgq=value2",
			want: "value1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryStrings(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return strings.Join(c.QueryStrings("fgq"), "|")
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=value1&fgq=value2",
			want: "value1|value2",
		},
		{
			name: "single value",
			url:  "/?fgq=Flamego&language=Go",
			want: "Flamego",
		},
		{
			name: "empty value",
			url:  "/?fgq=&language=Go",
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryEscape(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return c.QueryEscape("fgq")
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=%E4%B8%AD%E5%9B%BD%20666",
			want: "中国 666",
		},
		{
			name: "empty value",
			url:  "/?fgq=&language=Go",
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryBool(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return strconv.FormatBool(c.QueryBool("fgq"))
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=true",
			want: "true",
		},
		{
			name: "normal",
			url:  "/?fgq=False",
			want: "false",
		},
		{
			name: "empty value",
			url:  "/?fgq=",
			want: "false",
		},
		{
			name: "single char",
			url:  "/?fgq=T",
			want: "true",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryInt(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return strconv.FormatInt(int64(c.QueryInt("fgq")), 10)
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=123",
			want: "123",
		},
		{
			name: "empty value",
			url:  "/?fgq=",
			want: "0",
		},
		{
			name: "negative value",
			url:  "/?fgq=-123",
			want: "-123",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryInt64(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return strconv.FormatInt(c.QueryInt64("fgq"), 10)
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=123",
			want: "123",
		},
		{
			name: "empty value",
			url:  "/?fgq=",
			want: "0",
		},
		{
			name: "negative value",
			url:  "/?fgq=-123",
			want: "-123",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_QueryFloat64(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return fmt.Sprintf("%v", c.QueryFloat64("fgq"))
	})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "normal",
			url:  "/?fgq=3.1415926",
			want: "3.1415926",
		},
		{
			name: "empty value",
			url:  "/?fgq=",
			want: "0",
		},
		{
			name: "negative value",
			url:  "/?fgq=-3.1415926",
			want: "-3.1415926",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}

func TestContext_SetCookie(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) {
		c.SetCookie(
			http.Cookie{
				Name:  "country",
				Value: "中国",
				Path:  "/",
			},
		)
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "country=%E4%B8%AD%E5%9B%BD; Path=/", resp.Header().Get("Set-Cookie"))
}

func TestContext_Cookie(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Get("/", func(c Context) string {
		return c.Cookie("fgs")
	})

	tests := []struct {
		name   string
		cookie *http.Cookie
		want   string
	}{
		{
			name: "normal",
			cookie: &http.Cookie{
				Name:  "fgs",
				Value: "10086",
				Path:  "/",
			},
			want: "10086",
		},
		{
			name: "unescaped",
			cookie: &http.Cookie{
				Name:  "fgs",
				Value: "%E4%B8%AD%E5%9B%BD%20666",
				Path:  "/",
			},
			want: "中国 666",
		},
		{
			name: "not exists",
			cookie: &http.Cookie{
				Name:  "bad",
				Value: "10086",
				Path:  "/",
			},
			want: "",
		},
		{
			name: "unable to escape",
			cookie: &http.Cookie{
				Name:  "fgs",
				Value: "%E4%B%ADE5%9%BD%20666",
				Path:  "/",
			},
			want: "%E4%B%ADE5%9%BD%20666",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			assert.Nil(t, err)

			req.AddCookie(test.cookie)
			f.ServeHTTP(resp, req)

			assert.Equal(t, test.want, resp.Body.String())
		})
	}
}
