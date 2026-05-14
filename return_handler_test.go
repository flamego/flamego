// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type testReturnBody struct {
	text string
}

type testReturnError string

func (e testReturnError) Error() string { return string(e) }

func TestReturnHandler(t *testing.T) {
	tests := []struct {
		name     string
		handler  Handler
		wantCode int
		wantBody string
	}{
		{
			name: "(int, string)",
			handler: func() (int, string) {
				return http.StatusTeapot, "i'm a teapot"
			},
			wantCode: http.StatusTeapot,
			wantBody: "i'm a teapot",
		},
		{
			name: "(int, []byte)",
			handler: func() (int, []byte) {
				return http.StatusTeapot, []byte("i'm a teapot")
			},
			wantCode: http.StatusTeapot,
			wantBody: "i'm a teapot",
		},
		{
			name: "(int, error)",
			handler: func() (int, error) {
				return http.StatusForbidden, errors.New("teapot on the phone")
			},
			wantCode: http.StatusForbidden,
			wantBody: "teapot on the phone",
		},

		{
			name: "(string, error)",
			handler: func() (string, error) {
				return "", errors.New("teapot on the phone")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "teapot on the phone",
		},
		{
			name: "(string, nil-error)",
			handler: func() (string, error) {
				return "i'm a teapot", nil
			},
			wantCode: http.StatusOK,
			wantBody: "i'm a teapot",
		},
		{
			name: "([]byte, error)",
			handler: func() ([]byte, error) {
				return []byte(""), errors.New("teapot on the phone")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "teapot on the phone",
		},
		{
			name: "([]byte, nil-error)",
			handler: func() ([]byte, error) {
				return []byte("i'm a teapot"), nil
			},
			wantCode: http.StatusOK,
			wantBody: "i'm a teapot",
		},

		{
			name: "string",
			handler: func() string {
				return "my boss, my hero"
			},
			wantCode: http.StatusOK,
			wantBody: "my boss, my hero",
		},
		{
			name: "[]byte",
			handler: func() []byte {
				return []byte("my boss, my hero")
			},
			wantCode: http.StatusOK,
			wantBody: "my boss, my hero",
		},
		{
			name: "error",
			handler: func() error {
				return errors.New("teapot on the phone")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "teapot on the phone",
		},
		{
			name: "nil-error",
			handler: func() error {
				return nil
			},
			wantCode: http.StatusOK,
			wantBody: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := New()
			f.Get("/", test.handler)

			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)

			assert.Equal(t, test.wantCode, resp.Code)
			assert.Equal(t, test.wantBody, resp.Body.String())
		})
	}
}

func TestFlame_ReturnHandler(t *testing.T) {
	f := New()
	f.ReturnHandler(func(c Context, body testReturnBody) {
		_, _ = c.ResponseWriter().Write([]byte(body.text))
	})
	f.ReturnHandler(func(c Context, status int, body testReturnBody) {
		c.ResponseWriter().WriteHeader(status)
		_, _ = c.ResponseWriter().Write([]byte(body.text))
	})
	f.ReturnHandler(func(c Context, body testReturnBody, err error) {
		if err != nil {
			c.ResponseWriter().WriteHeader(http.StatusBadRequest)
			_, _ = c.ResponseWriter().Write([]byte(err.Error()))
			return
		}

		_, _ = c.ResponseWriter().Write([]byte(body.text))
	})

	f.Get("/body", func() testReturnBody {
		return testReturnBody{text: "body"}
	})
	f.Get("/status", func() (int, testReturnBody) {
		return http.StatusCreated, testReturnBody{text: "created"}
	})
	f.Get("/error", func() (testReturnBody, error) {
		return testReturnBody{}, errors.New("bad body")
	})
	f.Get("/nil-error", func() (testReturnBody, error) {
		return testReturnBody{text: "ok"}, nil
	})

	tests := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{path: "/body", wantCode: http.StatusOK, wantBody: "body"},
		{path: "/status", wantCode: http.StatusCreated, wantBody: "created"},
		{path: "/error", wantCode: http.StatusBadRequest, wantBody: "bad body"},
		{path: "/nil-error", wantCode: http.StatusOK, wantBody: "ok"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, test.path, nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)

			assert.Equal(t, test.wantCode, resp.Code)
			assert.Equal(t, test.wantBody, resp.Body.String())
		})
	}
}

func TestFlame_ReturnHandler_assignable(t *testing.T) {
	f := New()
	f.ReturnHandler(func(c Context, err error) {
		c.ResponseWriter().WriteHeader(http.StatusTeapot)
		_, _ = c.ResponseWriter().Write([]byte("error: " + err.Error()))
	})
	f.Get("/", func() testReturnError {
		return testReturnError("boom")
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusTeapot, resp.Code)
	assert.Equal(t, "error: boom", resp.Body.String())
}

func TestFlame_ReturnHandler_exactMatchBeforeAssignable(t *testing.T) {
	f := New()
	f.ReturnHandler(func(c Context, err error) {
		_, _ = c.ResponseWriter().Write([]byte("error: " + err.Error()))
	})
	f.ReturnHandler(func(c Context, err testReturnError) {
		_, _ = c.ResponseWriter().Write([]byte("exact: " + err.Error()))
	})
	f.Get("/", func() testReturnError {
		return testReturnError("boom")
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "exact: boom", resp.Body.String())
}

func TestFlame_ReturnHandler_register(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		f := New()
		assert.Panics(t, func() { f.ReturnHandler(nil) })
		assert.Panics(t, func() { f.ReturnHandler("string") })
		assert.Panics(t, func() { f.ReturnHandler(func() string { return "bad" }) })
		assert.Panics(t, func() { f.ReturnHandler(func(Context) {}) })
	})

	t.Run("duplicate", func(t *testing.T) {
		f := New()
		f.ReturnHandler(func(string) {})
		assert.Panics(t, func() {
			f.ReturnHandler(func(Context, string) {})
		})
	})
}
