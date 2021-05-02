// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTree(t *testing.T) {
	t.Run("empty segment", func(t *testing.T) {
		_, err := newTree(nil, &Segment{})
		got := fmt.Sprintf("%v", err)
		want := "empty segment in position 0"
		assert.Equal(t, want, got)
	})

	t.Run("empty segment element", func(t *testing.T) {
		_, err := newTree(nil, &Segment{Elements: []SegmentElement{{}}})
		got := fmt.Sprintf("%v", err)
		want := "empty segment element in position 0"
		assert.Equal(t, want, got)
	})

	parser, err := NewParser()
	assert.Nil(t, err)

	tests := []struct {
		route string
		style MatchStyle
		want  Tree
	}{
		{
			route: "/webapi/events",
			style: matchStyleStatic,
			want:  &staticTree{},
		},
		{
			route: "/{name}/events",
			style: matchStylePlaceholder,
			want:  &placeholderTree{},
		},
		{
			route: "/{paths: **}/events",
			style: matchStyleAll,
			want: &matchAllTree{
				bind: "paths",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)
			assert.Len(t, route.Segments, 2)

			segment := route.Segments[0]
			got, err := newTree(nil, segment)
			assert.Nil(t, err)

			switch test.style {
			case matchStyleStatic:
				test.want.(*staticTree).segment = segment
			case matchStylePlaceholder:
				test.want.(*placeholderTree).segment = segment
			case matchStyleAll:
				test.want.(*matchAllTree).segment = segment
			}

			assert.Equal(t, test.want, got)
		})
	}
}

func TestNewTree_Regex(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tests := []struct {
		route      string
		wantRegexp string
		wantBinds  []string
	}{
		{
			route:      "/{id: /[0-9]+/}/events",
			wantRegexp: `^([0-9]+)$`,
			wantBinds:  []string{"id"},
		},
		{
			route:      "/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}/events",
			wantRegexp: `^([0-9]{4})-([0-9]{2})-([0-9]{2})$`,
			wantBinds:  []string{"year", "month", "day"},
		},
		{
			route:      "/{hash: /[a-f0-9]{7,40}/}-{name}/events",
			wantRegexp: `^([a-f0-9]{7,40})-(.+)$`,
			wantBinds:  []string{"hash", "name"},
		},
		{
			route:      `/{before: /[a-z0-9]{40}/}...{after: /[a-z0-9]{40}/}/events`,
			wantRegexp: `^([a-z0-9]{40})\.\.\.([a-z0-9]{40})$`,
			wantBinds:  []string{"before", "after"},
		},
		{
			route:      `/article_{id: /[0-9]+/}_{page: /[\w]+/}.{ext: /diff|patch/}/events`,
			wantRegexp: `^article_([0-9]+)_([\w]+)\.(diff|patch)$`,
			wantBinds:  []string{"id", "page", "ext"},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)
			assert.Len(t, route.Segments, 2)

			segment := route.Segments[0]
			got, err := newTree(nil, segment)
			assert.Nil(t, err)

			tree := got.(*regexTree)
			assert.Equal(t, test.wantRegexp, tree.regexp.String())
			assert.Equal(t, test.wantBinds, tree.binds)
		})
	}
}

func TestTree_AddRoute(t *testing.T) {
	// todo
}
