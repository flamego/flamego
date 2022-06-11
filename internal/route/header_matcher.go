// Copyright 2022 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"regexp"
)

// HeaderMatcher stores matchers for request headers.
type HeaderMatcher struct {
	matches map[string]*regexp.Regexp // Key is the header name
}

// NewHeaderMatcher creates a new HeaderMatcher using given matches, where keys
// are header names.
func NewHeaderMatcher(matches map[string]*regexp.Regexp) *HeaderMatcher {
	return &HeaderMatcher{
		matches: matches,
	}
}

// Match returns true if all matches are successfully in the given header.
func (m *HeaderMatcher) Match(header http.Header) bool {
	for name, re := range m.matches {
		v := header.Get(name)
		if v == "" {
			return false
		}
		if !re.MatchString(v) {
			return false
		}
	}
	return true
}
