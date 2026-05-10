// Copyright 2026 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import "net/http"

// Predicate is an arbitrary request matching criterion. The route is matched
// only if the predicate returns true.
type Predicate func(*http.Request) bool

// PredicateMatcher stores a list of predicates that all must return true for
// the match to succeed.
type PredicateMatcher struct {
	predicates []Predicate
}

// NewPredicateMatcher creates a new PredicateMatcher with the given predicates.
func NewPredicateMatcher(predicates []Predicate) *PredicateMatcher {
	return &PredicateMatcher{
		predicates: predicates,
	}
}

// Match returns true if all predicates return true for the given request. A
// nil request is treated as a non-match when any predicate is configured.
func (m *PredicateMatcher) Match(req *http.Request) bool {
	if len(m.predicates) == 0 {
		return true
	}
	if req == nil {
		return false
	}
	for _, p := range m.predicates {
		if !p(req) {
			return false
		}
	}
	return true
}
