// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"regexp"
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
	parser, err := NewParser()
	assert.Nil(t, err)

	t.Run("duplicated match all style", func(t *testing.T) {
		route, err := parser.Parse(`/webapi/tree/{paths: **}/{names: **}/upload`)
		assert.Nil(t, err)

		_, err = NewTree().AddRoute(route, nil)
		got := fmt.Sprintf("%v", err)
		want := "new tree: duplicated match all style in position 25"
		assert.Equal(t, want, got)
	})

	tests := []struct {
		route     string
		style     MatchStyle
		wantDepth int
		wantLeaf  Leaf
	}{
		{
			route:     "/webapi",
			style:     matchStyleStatic,
			wantDepth: 2,
			wantLeaf: &staticLeaf{
				baseLeaf: baseLeaf{},
			},
		},
		{
			route:     "/webapi/name",
			style:     matchStyleStatic,
			wantDepth: 3,
			wantLeaf: &staticLeaf{
				baseLeaf: baseLeaf{},
			},
		},
		{
			route:     "/webapi/users/{name}",
			style:     matchStylePlaceholder,
			wantDepth: 4,
			wantLeaf: &placeholderLeaf{
				baseLeaf: baseLeaf{},
			},
		},
		{
			route:     "/webapi/tree/{paths: **}",
			style:     matchStyleAll,
			wantDepth: 4,
			wantLeaf: &matchAllLeaf{
				baseLeaf: baseLeaf{},
				bind:     "paths",
			},
		},
		{
			route:     "/webapi/tree/{paths: **}/edit",
			style:     matchStyleStatic,
			wantDepth: 5,
			wantLeaf: &staticLeaf{
				baseLeaf: baseLeaf{},
			},
		},
		{
			route:     "/webapi/tree/{paths: **}/edit/{name: **}",
			style:     matchStyleAll,
			wantDepth: 6,
			wantLeaf: &matchAllLeaf{
				baseLeaf: baseLeaf{},
				bind:     "name",
			},
		},
		{
			route:     `/webapi/article_{id: /[0-9]+/}_{page: /[\\w]+/}.{ext: /diff|patch/}`,
			style:     matchStyleRegex,
			wantDepth: 3,
			wantLeaf: &regexLeaf{
				baseLeaf: baseLeaf{},
				regexp:   regexp.MustCompile(`^article_([0-9]+)_([\\w]+)\.(diff|patch)$`),
				binds:    []string{"id", "page", "ext"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)

			got, err := NewTree().AddRoute(route, nil)
			assert.Nil(t, err)

			segment := route.Segments[len(route.Segments)-1]
			switch test.style {
			case matchStyleStatic:
				test.wantLeaf.(*staticLeaf).parent = got.Parent()
				test.wantLeaf.(*staticLeaf).segment = segment
				test.wantLeaf.(*staticLeaf).route = route
			case matchStyleRegex:
				test.wantLeaf.(*regexLeaf).parent = got.Parent()
				test.wantLeaf.(*regexLeaf).segment = segment
				test.wantLeaf.(*regexLeaf).route = route
			case matchStylePlaceholder:
				test.wantLeaf.(*placeholderLeaf).parent = got.Parent()
				test.wantLeaf.(*placeholderLeaf).segment = segment
				test.wantLeaf.(*placeholderLeaf).route = route
			case matchStyleAll:
				test.wantLeaf.(*matchAllLeaf).parent = got.Parent()
				test.wantLeaf.(*matchAllLeaf).segment = segment
				test.wantLeaf.(*matchAllLeaf).route = route
			}

			assert.Equal(t, test.wantLeaf, got)

			depth := 1
			ancestor := got.Parent()
			for ancestor != nil {
				ancestor = ancestor.getParent()
				depth++
			}
			assert.Equal(t, test.wantDepth, depth)
		})
	}
}
