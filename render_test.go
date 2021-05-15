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

func TestRender_JSON(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Use(Renderer())
	f.Get("/", func(r Render) {
		r.JSON(http.StatusUnauthorized, map[string]string{
			"status": http.StatusText(http.StatusUnauthorized),
		})
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status": "Unauthorized"}`, resp.Body.String())
}

func TestRender_XML(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Use(Renderer())
	f.Get("/", func(r Render) {
		type status struct {
			Code string `xml:"code"`
		}
		r.XML(http.StatusUnauthorized, status{
			Code: http.StatusText(http.StatusUnauthorized),
		})
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "text/xml; charset=utf-8", resp.Header().Get("Content-Type"))
	assert.Equal(t, `<status><code>Unauthorized</code></status>`, resp.Body.String())
}

func TestRender_Binary(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Use(Renderer())
	f.Get("/", func(r Render) {
		r.Binary(http.StatusUnauthorized, []byte{1, 2, 3})
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "application/octet-stream", resp.Header().Get("Content-Type"))
	assert.Equal(t, []byte{1, 2, 3}, resp.Body.Bytes())
}

func TestRender_PlainText(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Use(Renderer())
	f.Get("/", func(r Render) {
		r.PlainText(http.StatusUnauthorized, "hello world!")
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header().Get("Content-Type"))
	assert.Equal(t, `hello world!`, resp.Body.String())
}
