// Copyright 2026 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"testing"
)

var (
	benchmarkLeaf   Leaf
	benchmarkParams Params
	benchmarkOK     bool
)

func newBenchmarkTree(b *testing.B, routes ...string) Tree {
	b.Helper()

	parser, err := NewParser()
	if err != nil {
		b.Fatal(err)
	}

	tree := NewTree()
	for _, routePath := range routes {
		r, err := parser.Parse(routePath)
		if err != nil {
			b.Fatal(err)
		}

		_, err = AddRoute(tree, r, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	return tree
}

func BenchmarkTreeMatch(b *testing.B) {
	tree := newBenchmarkTree(b,
		"/webapi",
		"/webapi/users/?{id}",
		"/webapi/users/ids/{id: /[0-9]+/}",
		"/webapi/users/ids/{sha: /[a-z0-9]{7,40}/}",
		"/webapi/users/sessions/{paths: **}",
		"/webapi/users/events/{names: **}/feed",
		"/webapi/projects/{name}/hashes/{paths: **, capture: 2}/blob/{lineno: /[0-9]+/}",
		"/webapi/projects/{name}/commit/{sha: /[a-z0-9]{7,40}/}/main.go",
		`/webapi/projects/{name}/commit/{sha: /[a-z0-9]{7,40}/}{ext: /(\.(patch|diff))?/}`,
		"/webapi/articles/{category}/{year: /[0-9]{4}/}-{month}-{day}.json",
	)

	benchmarks := []struct {
		name string
		path string
		ok   bool
	}{
		{name: "static", path: "/webapi", ok: true},
		{name: "placeholder", path: "/webapi/users/12", ok: true},
		{name: "regex", path: "/webapi/articles/social/2021-05-03.json", ok: true},
		{name: "match_all_leaf", path: "/webapi/users/sessions/ab/cd/ef/gh", ok: true},
		{name: "match_all_tree", path: "/webapi/users/events/ab/cd/ef/gh/feed", ok: true},
		{name: "miss", path: "/webapi/projects/flamego/hashes/src/lib/blob/abc", ok: false},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			var leaf Leaf
			var params Params
			var ok bool
			for i := 0; i < b.N; i++ {
				leaf, params, ok = tree.Match(bm.path, nil)
				if ok != bm.ok {
					b.Fatalf("Match() ok = %v, want %v", ok, bm.ok)
				}
			}
			benchmarkLeaf = leaf
			benchmarkParams = params
			benchmarkOK = ok
		})
	}
}

func BenchmarkTreeMatchManyStaticSiblings(b *testing.B) {
	routes := make([]string, 0, 256)
	for i := 0; i < cap(routes); i++ {
		routes = append(routes, fmt.Sprintf("/webapi/repos/repo-%03d/{id}", i))
	}

	tree := newBenchmarkTree(b, routes...)
	b.ReportAllocs()

	var leaf Leaf
	var params Params
	var ok bool
	for i := 0; i < b.N; i++ {
		leaf, params, ok = tree.Match("/webapi/repos/repo-255/123", nil)
		if !ok {
			b.Fatal("Match() ok = false, want true")
		}
	}
	benchmarkLeaf = leaf
	benchmarkParams = params
	benchmarkOK = ok
}
