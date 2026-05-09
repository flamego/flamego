// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

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
			require.NoError(t, err)
			require.Len(t, route.Segments, 2)

			segment := route.Segments[0]
			got, err := newTree(nil, segment)
			require.NoError(t, err)

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
	require.NoError(t, err)

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
			require.NoError(t, err)
			require.Len(t, route.Segments, 2)

			segment := route.Segments[0]
			got, err := newTree(nil, segment)
			require.NoError(t, err)

			tree := got.(*regexTree)
			assert.Equal(t, test.wantRegexp, tree.regexp.String())
			assert.Equal(t, test.wantBinds, tree.binds)
		})
	}
}

func TestAddRoute(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	t.Run("duplicated routes", func(t *testing.T) {
		tree := NewTree()

		r1, err := parser.Parse(`/webapi/users`)
		require.NoError(t, err)
		_, err = AddRoute(tree, r1, nil)
		require.NoError(t, err)

		r2, err := parser.Parse(`/webapi/users/?events`)
		require.NoError(t, err)
		_, err = AddRoute(tree, r2, nil)
		got := fmt.Sprintf("%v", err)
		want := `add optional leaf to grandparent: duplicated route "/webapi/users/?events"`
		assert.Equal(t, want, got)
	})

	t.Run("adjacent unbounded match all styles", func(t *testing.T) {
		route, err := parser.Parse(`/webapi/tree/{paths: **}/{names: **}/upload`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		got := fmt.Sprintf("%v", err)
		want := "match all style in position 24 follows an unbounded match all style in position 12 with no separator; the preceding glob must have a capture limit"
		assert.Equal(t, want, got)
	})

	t.Run("unbounded match all followed later by another with no separator", func(t *testing.T) {
		// {a:**} is non-final and unbounded. {b:**} follows with no static/regex/placeholder anchor between them.
		route, err := parser.Parse(`/api/{a: **}/{b: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		got := fmt.Sprintf("%v", err)
		want := "match all style in position 12 follows an unbounded match all style in position 4 with no separator; the preceding glob must have a capture limit"
		assert.Equal(t, want, got)
	})

	t.Run("bounded then unbounded multi-glob is allowed", func(t *testing.T) {
		route, err := parser.Parse(`/api/{a: **, capture: 2}/{b: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		assert.NoError(t, err)
	})

	t.Run("static-separated multi-glob is allowed", func(t *testing.T) {
		route, err := parser.Parse(`/api/{a: **}/sep/{b: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		assert.NoError(t, err)
	})

	t.Run("unbounded then bounded multi-glob is rejected", func(t *testing.T) {
		// A capture limit on the *later* glob does not help. The earlier glob is
		// what consumes path segments first, and being unbounded it would swallow
		// what the later glob needs.
		route, err := parser.Parse(`/api/{a: **}/{b: **, capture: 2}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		got := fmt.Sprintf("%v", err)
		want := "match all style in position 12 follows an unbounded match all style in position 4 with no separator; the preceding glob must have a capture limit"
		assert.Equal(t, want, got)
	})

	t.Run("three globs with mixed separators is allowed", func(t *testing.T) {
		// Exercises the validation walk across a longer route: bounded, anchored
		// by static, bounded, anchored by static, unbounded final.
		route, err := parser.Parse(`/api/{a: **, capture: 2}/sep/{b: **, capture: 2}/end/{c: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		assert.NoError(t, err)
	})

	t.Run("two bounded globs followed by unbounded is allowed", func(t *testing.T) {
		// No separator is required between consecutive bounded globs. Each
		// capture limit prevents that glob from swallowing what follows, so the
		// trailing unbounded glob still has segments to match.
		route, err := parser.Parse(`/api/{a: **, capture: 2}/{b: **, capture: 2}/{c: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		assert.NoError(t, err)
	})

	t.Run("placeholder between unbounded globs is rejected", func(t *testing.T) {
		// A placeholder consumes exactly one segment of any content, so it does
		// not constrain how far the preceding unbounded glob can grow. Both
		// `a=x, id=y, b=z/w` and `a=x/y, id=z, b=w` would satisfy the pattern
		// for `/api/x/y/z/w`. The choice would depend on backtracking order
		// rather than on the route shape.
		route, err := parser.Parse(`/api/{a: **}/{id}/{b: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		got := fmt.Sprintf("%v", err)
		want := "match all style in position 17 follows an unbounded match all style in position 4 with no separator; the preceding glob must have a capture limit"
		assert.Equal(t, want, got)
	})

	t.Run("regex between unbounded globs is allowed", func(t *testing.T) {
		// A regex segment constrains the path text, so it acts as a true anchor
		// between globs.
		route, err := parser.Parse(`/api/{a: **}/{n: /[0-9]+/}/{b: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		assert.NoError(t, err)
	})

	t.Run("bounded then placeholder then unbounded is allowed", func(t *testing.T) {
		// The first glob has a capture limit, so the placeholder does not need
		// to separate two unbounded globs. Matching still follows the normal
		// bounded-glob partition rules.
		route, err := parser.Parse(`/api/{a: **, capture: 2}/{id}/{b: **}`)
		require.NoError(t, err)

		_, err = AddRoute(NewTree(), route, nil)
		assert.NoError(t, err)
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
			route:     `/webapi/article_{id: /\d+/}_{page: /[\\w]+/}.{ext: /diff|patch/}`,
			style:     matchStyleRegex,
			wantDepth: 3,
			wantLeaf: &regexLeaf{
				baseLeaf: baseLeaf{},
				regexp:   regexp.MustCompile(`^article_(\d+)_([\\w]+)\.(diff|patch)$`),
				binds:    []string{"id", "page", "ext"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			require.NoError(t, err)

			got, err := AddRoute(NewTree(), route, nil)
			require.NoError(t, err)

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
	require.NoError(t, err)

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
			require.NoError(t, err)

			_, err = AddRoute(tree, r, nil)
			got := fmt.Sprintf("%v", err)
			assert.Contains(t, got, "duplicated bind parameter")
		})
	}
}

func TestAddRoute_DuplicatedMatchAll(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

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
			require.NoError(t, err)

			_, err = AddRoute(tree, r, nil)
			if i%2 == 0 {
				assert.NoError(t, err)
			} else {
				got := fmt.Sprintf("%v", err)
				assert.Contains(t, got, "duplicated match all bind parameter")
			}
		})
	}
}

func TestTree_Match(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

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
		"/webapi/multi/{prefix: **, capture: 2}/{suffix: **}",
		// Sibling of the route above, sharing the "multi/{prefix: **, capture: 2}"
		// prefix subtree. At partition len=1 the static "sep" anchor matches. At
		// partition len=2 the match-all leaf above also matches. Match style
		// priority (static > matchAll) must override "longest partition wins".
		"/webapi/multi/{prefix: **, capture: 2}/sep/{y: **}",
		"/webapi/anchored/{a: **}/sep/{b: **}",
		// Three-level back-to-back bounded globs followed by an unbounded final
		// glob. No separators between them — each capture limit alone is enough
		// to keep the next glob from being starved.
		"/webapi/triple/{a: **, capture: 2}/{b: **, capture: 2}/{c: **}",
		"/webapi/cap1/{a: **, capture: 1}/{b: **}",
		"/webapi/strict/{a: **, capture: 2}/{n: /[0-9]+/}",
		// Two routes that share the same {leak: **, capture: 2} prefix and diverge below.
		// A failed partition descends into {x}/end, writing "x" before failing on "end".
		// The winning partition takes the regex sibling. If params weren't scratched
		// per trial, the stale "x" from the failed trial would leak into the result.
		"/webapi/leak/{leak: **, capture: 2}/{x}/end",
		"/webapi/leak/{leak: **, capture: 2}/{n: /[0-9]+/}",
		// Kitchen-sink route exercising most segment flavors in one shot:
		// statics, single-bind regex, bounded glob (non-final), placeholder,
		// multi-bind regex with literals between binds, and an unbounded final
		// glob. (Optional segments are excluded — they don't compose with a
		// trailing glob.)
		"/api/v1/{org_id: /[0-9]+/}/repos/{repo_path: **, capture: 3}/blob/{ref}/article_{id: /\\d+/}-{slug: /[a-z-]+/}.{ext: /md|txt/}/files/{tail: **}",
		"/webapi/special/test@$",
		"/webapi/special/%_",
	}
	for _, route := range routes {
		r, err := parser.Parse(route)
		require.NoError(t, err)

		_, err = AddRoute(tree, r, nil)
		require.NoError(t, err)
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
			// Bounded prefix glob (capture: 2) takes its maximum, suffix glob takes the rest.
			path:   "/webapi/multi/x/y/z/w",
			wantOK: true,
			wantParams: Params{
				"prefix": "x/y",
				"suffix": "z/w",
			},
		},
		{
			// Bounded prefix glob with fewer remaining segments still partitions greedily.
			path:   "/webapi/multi/x/y/z",
			wantOK: true,
			wantParams: Params{
				"prefix": "x/y",
				"suffix": "z",
			},
		},
		{
			// Sibling-priority guard for the bounded-glob greedy loop. Both
			// partitions succeed:
			//   len=1 via the static "sep" subtree → y=b/c
			//   len=2 via the match-all "{suffix}" leaf → suffix=b/c
			// The static-anchored branch is more specific and must win, even
			// though the match-all branch corresponds to a longer prefix
			// partition. If "longest wins" overrides priority, "y" would be
			// missing and "suffix" would be set instead.
			path:   "/webapi/multi/a/sep/b/c",
			wantOK: true,
			wantParams: Params{
				"prefix": "a",
				"y":      "b/c",
			},
		},
		{
			// Static separator anchors the first glob and matches the leftmost separator.
			path:   "/webapi/anchored/x/y/sep/z/w",
			wantOK: true,
			wantParams: Params{
				"a": "x/y",
				"b": "z/w",
			},
		},
		{
			// Three back-to-back globs, no separators. Each bounded glob takes
			// its maximum (capture: 2), and the unbounded final glob picks up
			// the remainder.
			path:   "/webapi/triple/x/y/z/w/v",
			wantOK: true,
			wantParams: Params{
				"a": "x/y",
				"b": "z/w",
				"c": "v",
			},
		},
		{
			// Same shape with fewer remaining segments: each bounded glob still
			// takes its maximum, the unbounded final glob takes a single segment.
			path:   "/webapi/triple/x/y/z/w",
			wantOK: true,
			wantParams: Params{
				"a": "x/y",
				"b": "z",
				"c": "w",
			},
		},
		{
			// capture: 1 — boundary case for the greedy-within-cap loop.
			path:   "/webapi/cap1/x/y/z",
			wantOK: true,
			wantParams: Params{
				"a": "x",
				"b": "y/z",
			},
		},
		{
			// Greedy loop must try multiple partitions: a=x with n=y fails (not digits),
			// a=x/y with n=42 succeeds. Proves the loop continues past failed downstream
			// matches up to the capture limit.
			path:   "/webapi/strict/x/y/42",
			wantOK: true,
			wantParams: Params{
				"a": "x/y",
				"n": "42",
			},
		},
		{
			// Param scratching: the leak=p partition tries the {x}/end branch, writes
			// x=q, then fails on "end". The leak=p/q partition then succeeds via the
			// regex sibling. The result must NOT contain a stale "x" from the failed
			// trial — only "leak" and "n".
			path:   "/webapi/leak/p/q/42",
			wantOK: true,
			wantParams: Params{
				"leak": "p/q",
				"n":    "42",
			},
		},
		{
			// Kitchen sink: every segment flavor binds correctly in one route.
			path:   "/api/v1/42/repos/a/b/c/blob/main/article_7-hello-world.md/files/x/y/z",
			wantOK: true,
			wantParams: Params{
				"org_id":    "42",
				"repo_path": "a/b/c",
				"ref":       "main",
				"id":        "7",
				"slug":      "hello-world",
				"ext":       "md",
				"tail":      "x/y/z",
			},
		},
		{
			// The bounded glob must continue past captured=1, where no "blob"
			// anchor matches, and commit at captured=2 when the static "blob"
			// subtree matches.
			path:   "/api/v1/42/repos/a/b/blob/main/article_7-hello-world.md/files/last",
			wantOK: true,
			wantParams: Params{
				"org_id":    "42",
				"repo_path": "a/b",
				"ref":       "main",
				"id":        "7",
				"slug":      "hello-world",
				"ext":       "md",
				"tail":      "last",
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
		{
			path:   "/webapi/anchored/x/y/z", // no "sep" segment to anchor the first glob
			wantOK: false,
		},
		{
			path:   "/webapi/strict/x/y/z/w", // no partition within capture: 2 leaves digits for "n"
			wantOK: false,
		},
		{
			// Kitchen sink, no match: "ext" must be "md" or "txt", not "html".
			path:   "/api/v1/42/repos/a/b/c/blob/main/article_7-hello-world.html/files/x",
			wantOK: false,
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			leaf, params, ok := tree.Match(test.path, nil)
			require.Equal(t, test.wantOK, ok)

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
	require.NoError(t, err)

	tree := NewTree()
	routes := []string{
		"/webapi/special/vars/{var}",
	}
	for _, route := range routes {
		r, err := parser.Parse(route)
		require.NoError(t, err)

		_, err = AddRoute(tree, r, nil)
		assert.NoError(t, err)
	}

	tests := []struct {
		path             string
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
			leaf, params, ok := tree.Match(test.path, nil)
			require.Equal(t, test.wantOK, ok)

			if !ok {
				return
			}
			assert.Equal(t, test.wantParams, params)
			assert.Equal(t, strings.TrimRight(test.wantUnescapedURL, "/"), leaf.URLPath(params, false))
		})
	}
}

func TestTree_MatchHeader(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	tree := NewTree()

	addRoute := func(path string, header map[string]string) {
		t.Helper()

		r, err := parser.Parse(path)
		require.NoError(t, err)

		l, err := AddRoute(tree, r, nil)
		assert.NoError(t, err)

		matches := make(map[string]*regexp.Regexp, len(header))
		for k, v := range header {
			matches[k] = regexp.MustCompile(v)
		}
		l.SetHeaderMatcher(NewHeaderMatcher(matches))
	}
	// Note: The order of routes and tests matters, matching for the same priority
	// is first in first match.
	addRoute("/webapi/static",
		map[string]string{
			"Server": "Caddy",
			"Status": "",
		},
	)

	addRoute("/webapi/vars/{var}",
		map[string]string{
			"Server": "Caddy",
		},
	)
	addRoute("/webapi/vars/{var}.html",
		map[string]string{
			"Server": "Caddy",
			"Status": "",
		},
	)

	addRoute(`/webapi/users/ids/{id: /[0-9]+/}_html`,
		map[string]string{
			"Server": "Caddy",
			"Status": "",
		},
	)
	addRoute(`/webapi/users/ids/{id: /\w+/}`,
		map[string]string{
			"Server": "Caddy",
		},
	)

	addRoute("/webapi/users/sessions/123",
		map[string]string{
			"Server": "Caddy",
			"Status": "",
		},
	)
	addRoute("/webapi/users/sessions/{paths: **}",
		map[string]string{
			"Server": "Caddy",
		},
	)

	addRoute("/webapi/users/events/{names: **}/feed",
		map[string]string{
			"Server": "Caddy",
			"Status": "",
		},
	)
	addRoute("/webapi/users/events/{names: **}",
		map[string]string{
			"Server": "Caddy",
		},
	)

	tests := []struct {
		path       string
		header     map[string]string
		wantOK     bool
		wantParams Params
	}{
		{
			path: "/webapi/static",
			header: map[string]string{
				"Server": "Caddy",
				"Status": "200 OK",
			},
			wantOK:     true,
			wantParams: Params{},
		},
		{
			path: "/webapi/static",
			header: map[string]string{
				"Server": "Caddy",
			},
			wantOK: false, // Missing "Status" header
		},

		{
			path: "/webapi/vars/abc.html",
			header: map[string]string{
				"Server": "Caddy",
			},
			wantOK: true,
			wantParams: Params{
				"var": "abc.html", // Not matching "/webapi/vars/{var}.html" because missing "Status" header
			},
		},
		{
			path: "/webapi/vars/abc.html",
			header: map[string]string{
				"Server": "Caddy",
				"Status": "200 OK",
			},
			wantOK: true,
			wantParams: Params{
				"var": "abc",
			},
		},

		{
			path: "/webapi/users/ids/abc_html",
			header: map[string]string{
				"Server": "Caddy",
			},
			wantOK: true,
			wantParams: Params{
				"id": "abc_html", // Not matching "/webapi/users/ids/{id: /[0-9]+/}_html" because missing "Status" header
			},
		},
		{
			path: "/webapi/users/ids/2830_html",
			header: map[string]string{
				"Server": "Caddy",
				"Status": "200 OK",
			},
			wantOK: true,
			wantParams: Params{
				"id": "2830",
			},
		},

		{
			path: "/webapi/users/sessions/123",
			header: map[string]string{
				"Server": "Caddy",
			},
			wantOK: true,
			wantParams: Params{
				"paths": "123", // Not matching "/webapi/users/sessions/123" because missing "Status" header
			},
		},
		{
			path: "/webapi/users/sessions/123",
			header: map[string]string{
				"Server": "Caddy",
				"Status": "200 OK",
			},
			wantOK:     true,
			wantParams: Params{},
		},

		{
			path: "/webapi/users/events/push/feed",
			header: map[string]string{
				"Server": "Caddy",
			},
			wantOK: true,
			wantParams: Params{
				"names": "push/feed", // Not matching "/webapi/users/events/{names: **}/feed" because missing "Status" header
			},
		},
		{
			path: "/webapi/users/events/push/feed",
			header: map[string]string{
				"Server": "Caddy",
				"Status": "200 OK",
			},
			wantOK: true,
			wantParams: Params{
				"names": "push",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			header := make(http.Header, len(test.header))
			for k, v := range test.header {
				header.Set(k, v)
			}
			leaf, params, ok := tree.Match(test.path, header)
			require.Equal(t, test.wantOK, ok)

			if !ok {
				return
			}
			assert.Equal(t, test.wantParams, params)
			assert.Equal(t, strings.TrimRight(test.path, "/"), leaf.URLPath(params, false))
		})
	}
}
