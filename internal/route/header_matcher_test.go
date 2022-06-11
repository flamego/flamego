// Copyright 2022 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderMatcher(t *testing.T) {
	header := make(http.Header)
	header.Set("Server", "Caddy")
	header.Set("Status", "200 OK")

	tests := []struct {
		name    string
		matches map[string]*regexp.Regexp
		want    bool
	}{
		{
			name: "loose matches",
			matches: map[string]*regexp.Regexp{
				"Server": regexp.MustCompile("Caddy"),
				"Status": regexp.MustCompile("200"),
			},
			want: true,
		},
		{
			name: "loose matches",
			matches: map[string]*regexp.Regexp{
				"Server": regexp.MustCompile("Caddy"),
				"Status": regexp.MustCompile("404"),
			},
			want: false,
		},

		{
			name: "exact matches",
			matches: map[string]*regexp.Regexp{
				"Server": regexp.MustCompile("^Caddy$"),
				"Status": regexp.MustCompile("^200 OK$"),
			},
			want: true,
		},
		{
			name: "exact matches",
			matches: map[string]*regexp.Regexp{
				"Server": regexp.MustCompile("^Caddy$"),
				"Status": regexp.MustCompile("^200$"),
			},
			want: false,
		},

		{
			name: "presence match",
			matches: map[string]*regexp.Regexp{
				"Server": regexp.MustCompile(""),
			},
			want: true,
		},
		{
			name: "presence match",
			matches: map[string]*regexp.Regexp{
				"Server":        regexp.MustCompile(""),
				"Cache-Control": regexp.MustCompile(""),
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewHeaderMatcher(test.matches).Match(header)
			assert.Equal(t, test.want, got)
		})
	}
}
