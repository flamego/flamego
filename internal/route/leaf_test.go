// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLeaf(t *testing.T) {
	t.Run("empty segment element", func(t *testing.T) {
		s := &Segment{Elements: []SegmentElement{{}}}
		_, err := newLeaf(nil, &Route{Segments: []*Segment{s}}, s, nil)
		got := fmt.Sprintf("%v", err)
		want := "empty segment element in position 0"
		assert.Equal(t, want, got)
	})

	parser, err := NewParser()
	assert.Nil(t, err)

	tests := []struct {
		route string
		style MatchStyle
		want  Leaf
	}{
		{
			route: "/webapi",
			style: matchStyleStatic,
			want: &staticLeaf{
				literals: "webapi",
			},
		},
		{
			route: "/{name}",
			style: matchStylePlaceholder,
			want: &placeholderLeaf{
				bind: "name",
			},
		},
		{
			route: "/{paths: **}",
			style: matchStyleAll,
			want: &matchAllLeaf{
				bind: "paths",
			},
		},
		{
			route: "/{**}",
			style: matchStyleAll,
			want: &matchAllLeaf{
				bind: "**",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)
			assert.Len(t, route.Segments, 1)

			segment := route.Segments[0]
			got, err := newLeaf(nil, route, segment, nil)
			assert.Nil(t, err)

			switch test.style {
			case matchStyleStatic:
				test.want.(*staticLeaf).segment = segment
				test.want.(*staticLeaf).route = route
			case matchStylePlaceholder:
				test.want.(*placeholderLeaf).segment = segment
				test.want.(*placeholderLeaf).route = route
			case matchStyleAll:
				test.want.(*matchAllLeaf).segment = segment
				test.want.(*matchAllLeaf).route = route
			}

			assert.Equal(t, test.want, got)
		})
	}
}

func TestNewLeaf_Regex(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tests := []struct {
		route      string
		wantRegexp string
		wantBinds  []string
	}{
		{
			route:      "/{id: /[0-9]+/}",
			wantRegexp: `^([0-9]+)$`,
			wantBinds:  []string{"id"},
		},
		{
			route:      "/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}",
			wantRegexp: `^([0-9]{4})-([0-9]{2})-([0-9]{2})$`,
			wantBinds:  []string{"year", "month", "day"},
		},
		{
			route:      "/{hash: /[a-f0-9]{7,40}/}-{name}",
			wantRegexp: `^([a-f0-9]{7,40})-(.+)$`,
			wantBinds:  []string{"hash", "name"},
		},
		{
			route:      `/{before: /[a-z0-9]{40}/}...{after: /[a-z0-9]{40}/}`,
			wantRegexp: `^([a-z0-9]{40})\.\.\.([a-z0-9]{40})$`,
			wantBinds:  []string{"before", "after"},
		},
		{
			route:      `/article_{id: /[0-9]+/}_{page: /[\w]+/}.{ext: /diff|patch/}`,
			wantRegexp: `^article_([0-9]+)_([\w]+)\.(diff|patch)$`,
			wantBinds:  []string{"id", "page", "ext"},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)
			assert.Len(t, route.Segments, 1)

			segment := route.Segments[0]
			got, err := newLeaf(nil, route, segment, nil)
			assert.Nil(t, err)

			leaf := got.(*regexLeaf)
			assert.Equal(t, test.wantRegexp, leaf.regexp.String())
			assert.Equal(t, test.wantBinds, leaf.binds)
		})
	}
}

func TestLeaf_URLPath(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tests := []struct {
		route        string
		vals         map[string]string
		withOptional bool
		want         string
	}{
		{
			route: "/webapi/users",
			vals: map[string]string{
				"404": "not found",
			},
			want: "/webapi/users",
		},
		{
			route: "/webapi/users/{name}",
			vals: map[string]string{
				"name": "alice",
			},
			want: "/webapi/users/alice",
		},
		{
			route: "/webapi/users/{name}/?events",
			vals: map[string]string{
				"name": "alice",
			},
			want: "/webapi/users/alice",
		},
		{
			route: "/webapi/users/{name}/?events",
			vals: map[string]string{
				"name": "alice",
			},
			withOptional: true,
			want:         "/webapi/users/alice/events",
		},
		{
			route: "/webapi/{paths: **}/files",
			vals: map[string]string{
				"paths": "src/lib",
			},
			want: "/webapi/src/lib/files",
		},
		{
			route: "/webapi/users/{id: /[0-9]+/}",
			vals: map[string]string{
				"id": "345",
			},
			want: "/webapi/users/345",
		},
		{
			route: "/webapi/posts/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}.html",
			vals: map[string]string{
				"year":  "2021",
				"month": "12",
				"day":   "24",
			},
			want: "/webapi/posts/2021-12-24.html",
		},
		{
			// NOTE: Purposely missing some values.
			route: "/webapi/posts/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}.html",
			vals: map[string]string{
				"year": "2021",
			},
			want: "/webapi/posts/2021-{month}-{day}.html",
		},
		{
			// NOTE: Purposely having unused some values.
			route: "/webapi/posts/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}.html",
			vals: map[string]string{
				"year":  "2021",
				"month": "12",
				"day":   "24",
				"time":  "12:11 PM",
			},
			want: "/webapi/posts/2021-12-24.html",
		},
		{
			route: `/webapi/compare/{before: /[a-z0-9]{40}/}...{after: /[a-z0-9]{40}/}`,
			vals: map[string]string{
				"before": "9aac00eb28cb0f04740ac75e69d85a917a292266",
				"after":  "74a6e8d74767fcb19b36f342203c5cb678031bb3",
			},
			want: "/webapi/compare/9aac00eb28cb0f04740ac75e69d85a917a292266...74a6e8d74767fcb19b36f342203c5cb678031bb3",
		},
		{
			route: `/webapi/article_{id: /[0-9]+/}_{page: /[\w]+/}.{ext: /diff|patch/}`,
			vals: map[string]string{
				"id":   "123",
				"page": "helloworld",
				"ext":  "diff",
			},
			want: "/webapi/article_123_helloworld.diff",
		},
		{
			route: `/webapi/{username}/%E4%BD%A0%E5%A5%BD%E4%B8%96%E7%95%8C/test@$`,
			vals: map[string]string{
				"username": "@hello",
			},
			want: "/webapi/@hello/%E4%BD%A0%E5%A5%BD%E4%B8%96%E7%95%8C/test@$",
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)

			segment := route.Segments[len(route.Segments)-1]
			leaf, err := newLeaf(nil, route, segment, nil)
			assert.Nil(t, err)

			got := leaf.URLPath(test.vals, test.withOptional)
			assert.Equal(t, test.want, got)
		})
	}
}
