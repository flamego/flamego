// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponseWriter(t *testing.T) {
	t.Run("write string", func(t *testing.T) {
		resp := httptest.NewRecorder()
		w := NewResponseWriter(http.MethodGet, resp)
		assert.False(t, w.Written())

		_, _ = w.Write([]byte("Hello world"))

		assert.Equal(t, w.Status(), resp.Code)
		assert.Equal(t, "Hello world", resp.Body.String())
		assert.Equal(t, http.StatusOK, w.Status())
		assert.Equal(t, 11, w.Size())
		assert.True(t, w.Written())
	})

	t.Run("write strings", func(t *testing.T) {
		resp := httptest.NewRecorder()
		w := NewResponseWriter(http.MethodGet, resp)
		assert.False(t, w.Written())

		_, _ = w.Write([]byte("Hello world"))
		_, _ = w.Write([]byte("foo bar bat baz"))

		assert.Equal(t, w.Status(), resp.Code)
		assert.Equal(t, "Hello worldfoo bar bat baz", resp.Body.String())
		assert.Equal(t, http.StatusOK, w.Status())
		assert.Equal(t, 26, w.Size())
		assert.True(t, w.Written())
	})

	t.Run("write header", func(t *testing.T) {
		resp := httptest.NewRecorder()
		w := NewResponseWriter(http.MethodGet, resp)
		assert.False(t, w.Written())

		w.WriteHeader(http.StatusNotFound)

		assert.Equal(t, w.Status(), resp.Code)
		assert.Empty(t, resp.Body.String())
		assert.Equal(t, http.StatusNotFound, w.Status())
		assert.Equal(t, 0, w.Size())
		assert.True(t, w.Written())
	})

	t.Run("before funcs", func(t *testing.T) {
		resp := httptest.NewRecorder()
		w := NewResponseWriter(http.MethodGet, resp)
		assert.False(t, w.Written())

		var buf bytes.Buffer
		w.Before(func(ResponseWriter) {
			buf.WriteString("foo")
		})
		w.Before(func(ResponseWriter) {
			buf.WriteString("bar")
		})
		w.WriteHeader(http.StatusNotFound)

		assert.Equal(t, w.Status(), resp.Code)
		assert.Empty(t, resp.Body.String())
		assert.Equal(t, http.StatusNotFound, w.Status())
		assert.Equal(t, 0, w.Size())

		assert.Equal(t, "barfoo", buf.String())
	})
}

type hijackableResponse struct {
	Hijacked bool
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{}
}

func (*hijackableResponse) Header() http.Header       { return nil }
func (*hijackableResponse) Write([]byte) (int, error) { return 0, nil }
func (*hijackableResponse) WriteHeader(int)           {}
func (*hijackableResponse) Flush()                    {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func TestResponseWriter_Hijack(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		hijackable := newHijackableResponse()
		w := NewResponseWriter(http.MethodGet, hijackable)

		hijacker, ok := w.(http.Hijacker)
		assert.True(t, ok)

		_, _, err := hijacker.Hijack()
		assert.Nil(t, err)
		assert.True(t, hijackable.Hijacked)
	})

	t.Run("bad", func(t *testing.T) {
		hijackable := new(http.ResponseWriter)
		rw := NewResponseWriter(http.MethodGet, *hijackable)

		hijacker, ok := rw.(http.Hijacker)
		assert.True(t, ok)

		_, _, err := hijacker.Hijack()
		assert.NotNil(t, err)
	})
}

func TestResponseWriter_Flush(t *testing.T) {
	resp := httptest.NewRecorder()
	w := NewResponseWriter(http.MethodGet, resp)

	f, ok := w.(http.Flusher)
	assert.True(t, ok)

	r := newRouter(newContext)
	r.Get("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for i := 0; i < 2; i++ {
			time.Sleep(10 * time.Millisecond)
			_, _ = io.WriteString(w, "data: Hello\n\n")
			f.Flush()
		}
	})

	req, err := http.NewRequest(http.MethodGet, "/events", nil)
	assert.Nil(t, err)

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, "data: Hello\n\ndata: Hello\n\n", resp.Body.String())
}

func TestResponseWriter_Push(t *testing.T) {
	resp := httptest.NewRecorder()
	w := NewResponseWriter(http.MethodGet, resp)

	_, ok := w.(http.Pusher)
	assert.True(t, ok)
}
