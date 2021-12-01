// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"regexp"
	"strings"
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
			want: &placeholderTree{
				bind: "name",
			},
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

func TestAddRoute(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	t.Run("duplicated routes", func(t *testing.T) {
		tree := NewTree()

		r1, err := parser.Parse(`/webapi/users`)
		assert.Nil(t, err)
		_, err = AddRoute(tree, r1, nil)
		assert.Nil(t, err)

		r2, err := parser.Parse(`/webapi/users/?events`)
		assert.Nil(t, err)
		_, err = AddRoute(tree, r2, nil)
		got := fmt.Sprintf("%v", err)
		want := `add optional leaf to grandparent: duplicated route "/webapi/users/?events"`
		assert.Equal(t, want, got)
	})

	t.Run("duplicated match all styles", func(t *testing.T) {
		route, err := parser.Parse(`/webapi/tree/{paths: **}/{names: **}/upload`)
		assert.Nil(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		got := fmt.Sprintf("%v", err)
		want := "new tree: duplicated match all style in position 24"
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
				literals: "webapi",
			},
		},
		{
			route:     "/webapi/name",
			style:     matchStyleStatic,
			wantDepth: 3,
			wantLeaf: &staticLeaf{
				baseLeaf: baseLeaf{},
				literals: "name",
			},
		},
		{
			route:     "/webapi/users/{name}",
			style:     matchStylePlaceholder,
			wantDepth: 4,
			wantLeaf: &placeholderLeaf{
				baseLeaf: baseLeaf{},
				bind:     "name",
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
				literals: "edit",
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

			got, err := AddRoute(NewTree(), route, nil)
			assert.Nil(t, err)

			segment := route.Segments[len(route.Segments)-1]
			switch test.style {
			case matchStyleStatic:
				test.wantLeaf.(*staticLeaf).parent = got.getParent()
				test.wantLeaf.(*staticLeaf).segment = segment
				test.wantLeaf.(*staticLeaf).route = route
			case matchStyleRegex:
				test.wantLeaf.(*regexLeaf).parent = got.getParent()
				test.wantLeaf.(*regexLeaf).segment = segment
				test.wantLeaf.(*regexLeaf).route = route
			case matchStylePlaceholder:
				test.wantLeaf.(*placeholderLeaf).parent = got.getParent()
				test.wantLeaf.(*placeholderLeaf).segment = segment
				test.wantLeaf.(*placeholderLeaf).route = route
			case matchStyleAll:
				test.wantLeaf.(*matchAllLeaf).parent = got.getParent()
				test.wantLeaf.(*matchAllLeaf).segment = segment
				test.wantLeaf.(*matchAllLeaf).route = route
			}

			assert.Equal(t, test.wantLeaf, got)

			depth := 1
			ancestor := got.getParent()
			for ancestor != nil {
				ancestor = ancestor.getParent()
				depth++
			}
			assert.Equal(t, test.wantDepth, depth)
		})
	}
}

func TestAddRoute_DuplicatedBinds(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tree := NewTree()

	routes := []string{
		// Leaf
		"/webapi/{name}/{name}",
		"/webapi/{name}/{name: /regexp/}",
		"/webapi/{name}/{name: **}",

		// Tree
		"/webapi/{name}/{name}/events",
		"/webapi/{name}/{name: /regexp/}/events",
		"/webapi/{name}/{name: **}/events",
	}
	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			r, err := parser.Parse(route)
			assert.Nil(t, err)

			_, err = AddRoute(tree, r, nil)
			got := fmt.Sprintf("%v", err)
			assert.Contains(t, got, "duplicated bind parameter")
		})
	}
}

func TestAddRoute_DuplicatedMatchAll(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tree := NewTree()

	routes := []string{
		// Leaf
		"/webapi/{name: **}",
		"/webapi/{user: **}",

		// Tree
		"/webapi/{name: **}/events",
		"/webapi/{user: **}/events",
	}
	for i, route := range routes {
		t.Run(route, func(t *testing.T) {
			r, err := parser.Parse(route)
			assert.Nil(t, err)

			_, err = AddRoute(tree, r, nil)

			if i%2 == 0 {
				assert.Nil(t, err)
			} else {
				got := fmt.Sprintf("%v", err)
				assert.Contains(t, got, "duplicated match all bind parameter")
			}
		})
	}
}

func TestTree_Match(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tree := NewTree()

	// NOTE: The order of routes and tests matters, matching for the same priority
	//  is first in first match.
	routes := []string{
		"/webapi",
		"/webapi/users/?{id}",
		"/webapi/users/ids/{id: /[0-9]+/}",
		"/webapi/users/ids/{sha: /[a-z0-9]{7,40}/}",
		"/webapi/users/sessions/{paths: **}",
		"/webapi/users/events/{names: **}/feed",
		"/webapi/users/settings/?profile",
		"/webapi/projects/{name}/hashes/{paths: **, capture: 2}/blob/{lineno: /[0-9]+/}",
		"/webapi/projects/{name}/commit/{sha: /[a-z0-9]{7,40}/}/main.go",
		`/webapi/projects/{name}/commit/{sha: /[a-z0-9]{7,40}/}{ext: /(\.(patch|diff))?/}`,
		"/webapi/articles/{category}/{year: /[0-9]{4}/}-{month}-{day}.json",
		"/webapi/groups/{name: **, capture: 2}",
		"/webapi/special/test@$",
		"/webapi/special/%_",
	}
	for _, route := range routes {
		r, err := parser.Parse(route)
		assert.Nil(t, err)

		_, err = AddRoute(tree, r, nil)
		assert.Nil(t, err)
	}

	tests := []struct {
		path         string
		withOptional bool
		wantOK       bool
		wantParams   Params
	}{
		// Match
		{
			path:       "/webapi",
			wantOK:     true,
			wantParams: Params{},
		},
		{
			path:       "/webapi/users",
			wantOK:     true,
			wantParams: Params{},
		},
		{
			path:         "/webapi/users/12",
			withOptional: true,
			wantOK:       true,
			wantParams: Params{
				"id": "12",
			},
		},
		{
			path:   "/webapi/users/ids/123", // Matched before "/webapi/users/ids/{sha: /[a-z0-9]{7,40}/}"
			wantOK: true,
			wantParams: Params{
				"id": "123",
			},
		},
		{
			path:   "/webapi/users/ids/368c7b2d0b1e0b243b2",
			wantOK: true,
			wantParams: Params{
				"sha": "368c7b2d0b1e0b243b2",
			},
		},
		{
			path:   "/webapi/users/sessions/ab/cd/ef/gh",
			wantOK: true,
			wantParams: Params{
				"paths": "ab/cd/ef/gh",
			},
		},
		{
			path:   "/webapi/users/events/ab/cd/ef/gh/feed",
			wantOK: true,
			wantParams: Params{
				"names": "ab/cd/ef/gh",
			},
		},
		{
			path:   "/webapi/projects/flamego/hashes/src/lib/blob/15",
			wantOK: true,
			wantParams: Params{
				"name":   "flamego",
				"paths":  "src/lib",
				"lineno": "15",
			},
		},
		{
			path:   "/webapi/projects/flamego/commit/368c7b2d0b1e0b243b2/main.go",
			wantOK: true,
			wantParams: Params{
				"name": "flamego",
				"sha":  "368c7b2d0b1e0b243b2",
			},
		},
		{
			path:   "/webapi/projects/flamego/commit/368c7b2d0b1e0b243b2", // "ext" is optional
			wantOK: true,
			wantParams: Params{
				"name": "flamego",
				"sha":  "368c7b2d0b1e0b243b2",
				"ext":  "",
			},
		},
		{
			path:   "/webapi/projects/flamego/commit/368c7b2d0b1e0b243b2.patch",
			wantOK: true,
			wantParams: Params{
				"name": "flamego",
				"sha":  "368c7b2d0b1e0b243b2",
				"ext":  ".patch",
			},
		},
		{
			path:   "/webapi/articles/social/2021-05-03.json",
			wantOK: true,
			wantParams: Params{
				"category": "social",
				"year":     "2021",
				"month":    "05",
				"day":      "03",
			},
		},
		{
			path:   "/webapi/groups/flamego/flamego",
			wantOK: true,
			wantParams: Params{
				"name": "flamego/flamego",
			},
		},
		{
			path:       "/webapi/special/test@$",
			wantOK:     true,
			wantParams: Params{},
		},
		{
			path:       "/webapi/special/%_",
			wantOK:     true,
			wantParams: Params{},
		},
		{
			path:       "/webapi/users/settings",
			wantOK:     true,
			wantParams: Params{},
		},
		{
			path:         "/webapi/users/settings/profile",
			withOptional: true,
			wantOK:       true,
			wantParams:   Params{},
		},

		// No match
		{
			path:   "/webapi//", // the last slash is a new route segment
			wantOK: false,
		},
		{
			path:   "/webapi/users/ids/abc", // "abc" are not digits
			wantOK: false,
		},
		{
			path:   "/webapi/projects/flamego/hashes/src/lib/blob/abc", // "abc" are not digits
			wantOK: false,
		},
		{
			path:   "/webapi/projects/flamego/commit/368c7b/main.go", // "368c7b" is less than 7 chars
			wantOK: false,
		},
		{
			path:   "/webapi/articles/social/21-05-03.json", // "21" length is not 4
			wantOK: false,
		},
		{
			path:   "/webapi/articles/social/year-05-03.json", // "year" are not digits
			wantOK: false,
		},
		{
			path:   "/webapi/articles/social/2021-05.json", // "day" is missing
			wantOK: false,
		},
		{
			path:   "/webapi/groups/flamego/flamego/flamego", // capture limit is 2
			wantOK: false,
		},
		{
			path:   "/webapi/projects/flamego/hashes/src/lib/main.c/blob/15", // capture limit is 2
			wantOK: false,
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			leaf, params, ok := tree.Match(test.path)
			assert.Equal(t, test.wantOK, ok)

			if !ok {
				return
			}
			assert.Equal(t, test.wantParams, params)
			assert.Equal(t, strings.TrimRight(test.path, "/"), leaf.URLPath(params, test.withOptional))
		})
	}
}

func TestTree_MatchEscape(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	tree := NewTree()

	// NOTE: The order of routes and tests matters, matching for the same priority
	//  is first in first match.
	routes := []string{
		"/webapi/special/vars/{var}",
	}
	for _, route := range routes {
		r, err := parser.Parse(route)
		assert.Nil(t, err)

		_, err = AddRoute(tree, r, nil)
		assert.Nil(t, err)
	}

	tests := []struct {
		path             string
		withOptional     bool
		wantOK           bool
		wantParams       Params
		wantUnescapedURL string
	}{
		{
			path:   "/webapi/special/vars/%_",
			wantOK: true,
			wantParams: Params{
				"var": "%_",
			},
			wantUnescapedURL: "/webapi/special/vars/%_",
		},
		{
			path:   "/webapi/special/vars/%E4%BD%A0%E5%A5%BD%E4%B8%96%E7%95%8C",
			wantOK: true,
			wantParams: Params{
				"var": "你好世界",
			},
			wantUnescapedURL: "/webapi/special/vars/你好世界",
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			leaf, params, ok := tree.Match(test.path)
			assert.Equal(t, test.wantOK, ok)

			if !ok {
				return
			}
			assert.Equal(t, test.wantParams, params)
			assert.Equal(t, strings.TrimRight(test.wantUnescapedURL, "/"), leaf.URLPath(params, test.withOptional))
		})
	}
}
