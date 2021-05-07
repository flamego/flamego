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

func TestRecovery(t *testing.T) {
	t.Run("recovery from panic", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Recovery())
		f.Use(func() { panic("here is a panic!") })
		f.Get("/", func() {})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "text/html", resp.Header().Get("Content-Type"))
		assert.Contains(t, resp.Body.String(), "PANIC")
	})

	t.Run("recovery from panic in non-development mode", func(t *testing.T) {
		SetEnv(EnvTypeProd)
		defer SetEnv(EnvTypeDev)

		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Recovery())
		f.Use(func() { panic("here is a panic!") })
		f.Get("/", func() {})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "text/plain", resp.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusText(http.StatusInternalServerError), resp.Body.String())
	})

	t.Run("Revocery panic to another ResponseWriter", func(t *testing.T) {
		resp := httptest.NewRecorder()
		resp2 := httptest.NewRecorder()

		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Recovery())
		f.Use(func(c Context) {
			c.MapTo(resp2, (*http.ResponseWriter)(nil))
			panic("here is a panic!")
		})
		f.Get("/", func() {})

		req, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp2.Code)
		assert.Equal(t, "text/html", resp2.Header().Get("Content-Type"))
		assert.Contains(t, resp2.Body.String(), "PANIC")
	})
}
