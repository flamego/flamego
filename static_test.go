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
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	f := NewWithLogger(&bytes.Buffer{})
	f.Use(Static())

	t.Run("serve with GET", func(t *testing.T) {
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/.editorconfig", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Empty(t, resp.Header().Get("Expires"))
		assert.NotEmpty(t, resp.Body.String())
	})

	t.Run("serve with HEAD", func(t *testing.T) {
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/.editorconfig", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Empty(t, resp.Header().Get("Expires"))
		assert.Empty(t, resp.Body.String())
	})

	t.Run("404 with POST", func(t *testing.T) {
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/.editorconfig", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}

func TestStatic_Options(t *testing.T) {
	t.Run("custom FileSystem", func(t *testing.T) {
		mockFS := &fstest.MapFS{
			"hello.txt": {
				Data: []byte("hello, world"),
			},
		}
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				FileSystem: http.FS(mockFS),
			},
		))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/hello.txt", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "hello, world", resp.Body.String())
	})

	t.Run("prefix", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				Prefix: "public",
			},
		))

		tests := []struct {
			url            string
			wantStatusCode int
		}{
			{
				url:            "/public/.editorconfig",
				wantStatusCode: http.StatusOK,
			},
			{
				url:            "/.editorconfig",
				wantStatusCode: http.StatusNotFound,
			},
		}
		for _, test := range tests {
			t.Run(test.url, func(t *testing.T) {
				resp := httptest.NewRecorder()
				req, err := http.NewRequest("HEAD", test.url, nil)
				assert.Nil(t, err)

				f.ServeHTTP(resp, req)

				assert.Equal(t, test.wantStatusCode, resp.Code)
			})
		}
	})

	t.Run("index", func(t *testing.T) {
		tests := []struct {
			url            string
			index          string
			wantStatusCode int
		}{
			{
				url:            "/",
				index:          ".editorconfig",
				wantStatusCode: http.StatusOK,
			},
			{
				url:            "/",
				index:          "index.html",
				wantStatusCode: http.StatusNotFound,
			},
		}
		for _, test := range tests {
			t.Run(test.index, func(t *testing.T) {
				f := NewWithLogger(&bytes.Buffer{})
				f.Use(Static(
					StaticOptions{
						Index: test.index,
					},
				))

				resp := httptest.NewRecorder()
				req, err := http.NewRequest("GET", test.url, nil)
				assert.Nil(t, err)

				f.ServeHTTP(resp, req)

				assert.Equal(t, test.wantStatusCode, resp.Code)
				fmt.Println(resp.Body.String())
			})
		}
	})

	t.Run("expires", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				Expires: func() string {
					return "2830"
				},
			},
		))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/.editorconfig", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "2830", resp.Header().Get("Expires"))
	})

	t.Run("etag", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				SetETag: true,
			},
		))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/.editorconfig", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		lastModified, err := time.Parse(http.TimeFormat, resp.Header().Get("Last-Modified"))
		assert.Nil(t, err)

		etag := generateETag(int64(resp.Body.Len()), ".editorconfig", lastModified)
		assert.Equal(t, etag, resp.Header().Get("ETag"))
	})

	t.Run("etag with If-None-Match", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				SetETag: true,
			},
		))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/.editorconfig", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		lastModified, err := time.Parse(http.TimeFormat, resp.Header().Get("Last-Modified"))
		assert.Nil(t, err)
		etag := generateETag(int64(resp.Body.Len()), ".editorconfig", lastModified)

		// Second request with ETag in If-None-Match
		resp = httptest.NewRecorder()
		req.Header.Add("If-None-Match", etag)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotModified, resp.Code)
	})
}

func TestStatic_Redirect(t *testing.T) {
	t.Run("serve with prefix but without redirect", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				Prefix: "/public",
			},
		))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/public/", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("serve with redirect", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static(
			StaticOptions{
				Prefix: "/public",
			},
		))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/public", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)
		assert.Equal(t, "/public/", resp.Header().Get("Location"))
	})

	t.Run("serve with improper request", func(t *testing.T) {
		f := NewWithLogger(&bytes.Buffer{})
		f.Use(Static())

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost:2830//example.com%2f..", nil)
		assert.Nil(t, err)

		f.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}
