// Copyright 2026 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPredicateMatcher(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	tests := []struct {
		name       string
		predicates []Predicate
		req        *http.Request
		want       bool
	}{
		{
			name:       "no predicates",
			predicates: nil,
			req:        req,
			want:       true,
		},
		{
			name:       "no predicates with nil request",
			predicates: nil,
			req:        nil,
			want:       true,
		},
		{
			name: "all true",
			predicates: []Predicate{
				func(*http.Request) bool { return true },
				func(*http.Request) bool { return true },
			},
			req:  req,
			want: true,
		},
		{
			name: "first false short-circuits",
			predicates: []Predicate{
				func(*http.Request) bool { return false },
				func(*http.Request) bool { panic("must not be called") },
			},
			req:  req,
			want: false,
		},
		{
			name: "last false",
			predicates: []Predicate{
				func(*http.Request) bool { return true },
				func(*http.Request) bool { return false },
			},
			req:  req,
			want: false,
		},
		{
			name: "nil request with predicates fails",
			predicates: []Predicate{
				func(*http.Request) bool { return true },
			},
			req:  nil,
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewPredicateMatcher(test.predicates).Match(test.req)
			assert.Equal(t, test.want, got)
		})
	}
}
