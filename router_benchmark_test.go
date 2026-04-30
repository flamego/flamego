// Copyright 2026 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"io"
	"net/http"
	"testing"
)

type benchmarkResponseWriter struct {
	header http.Header
}

func (w *benchmarkResponseWriter) Header() http.Header {
	return w.header
}

func (*benchmarkResponseWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (*benchmarkResponseWriter) WriteHeader(int) {}

func BenchmarkRouterServeHTTP(b *testing.B) {
	f := NewWithLogger(io.Discard)
	f.Get("/static", func() {})
	f.Get("/users/{id}", func(c Context) {
		_ = c.Param("id")
	})

	benchmarks := []struct {
		name string
		req  *http.Request
	}{
		{name: "static", req: mustBenchmarkRequest(b, http.MethodGet, "/static")},
		{name: "dynamic", req: mustBenchmarkRequest(b, http.MethodGet, "/users/123")},
		{name: "miss", req: mustBenchmarkRequest(b, http.MethodGet, "/missing")},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			w := &benchmarkResponseWriter{header: make(http.Header)}
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				f.ServeHTTP(w, bm.req)
			}
		})
	}
}

func mustBenchmarkRequest(b *testing.B, method, target string) *http.Request {
	b.Helper()

	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		b.Fatal(err)
	}
	return req
}
