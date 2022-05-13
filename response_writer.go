// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bufio"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// ResponseWriter is a wrapper of http.ResponseWriter that provides extra
// information about the response. It is recommended that middleware handlers
// use this construct to wrap a ResponseWriter if the functionality calls for
// it.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Pusher
	// Status returns the status code of the response or 0 if the response has not
	// been written.
	Status() int
	// Written returns whether the ResponseWriter has been written.
	Written() bool
	// Size returns the written size of the response body.
	Size() int
	// Before allows for a function to be called before the ResponseWriter has been
	// written to. This is useful for setting headers or any other operations that
	// must happen before a response has been written. Multiple calls to this method
	// will stack up functions, and functions will be called in the FILO manner.
	Before(BeforeFunc)
}

type responseWriter struct {
	http.ResponseWriter

	method      string       // The HTTP method of the coming request.
	status      int32        // The written status of the response.
	size        int          // The written size of the response.
	beforeFuncs []BeforeFunc // The list of functions to be called before written to the response.

	writeHeaderOnce sync.Once
}

// BeforeFunc is a function that is called before the ResponseWriter is written.
type BeforeFunc func(ResponseWriter)

// NewResponseWriter returns a wrapper of http.ResponseWriter.
func NewResponseWriter(method string, w http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		ResponseWriter: w,
		method:         method,
	}
}

func (w *responseWriter) callBefore() {
	for i := len(w.beforeFuncs) - 1; i >= 0; i-- {
		w.beforeFuncs[i](w)
	}
}

func (w *responseWriter) WriteHeader(s int) {
	w.writeHeaderOnce.Do(func() {
		if w.Written() {
			return
		}

		w.callBefore()
		w.ResponseWriter.WriteHeader(s)
		atomic.StoreInt32(&w.status, int32(s))
	})
}

func (w *responseWriter) Write(b []byte) (size int, err error) {
	if !w.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet.
		w.WriteHeader(http.StatusOK)
	}
	if w.method != http.MethodHead {
		size, err = w.ResponseWriter.Write(b)
		w.size += size
	}
	return size, err
}

func (w *responseWriter) Status() int {
	return int(atomic.LoadInt32(&w.status))
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.Status() != 0
}

func (w *responseWriter) Before(before BeforeFunc) {
	w.beforeFuncs = append(w.beforeFuncs, before)
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (w *responseWriter) Flush() {
	if !w.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet.
		w.WriteHeader(http.StatusOK)
	}

	flusher, ok := w.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return errors.New("the ResponseWriter doesn't support the Pusher interface")
	}
	return pusher.Push(target, opts)
}
