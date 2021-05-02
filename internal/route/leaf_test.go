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
	t.Run("empty segment", func(t *testing.T) {
		_, err := newLeaf(nil, &Segment{}, nil)
		got := fmt.Sprintf("%v", err)
		want := "empty segment in position 0"
		assert.Equal(t, want, got)
	})

	t.Run("empty segment element", func(t *testing.T) {
		_, err := newLeaf(nil, &Segment{Elements: []SegmentElement{{}}}, nil)
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
			want:  &staticLeaf{},
		},
		{
			route: "/{name}",
			style: matchStylePlaceholder,
			want:  &placeholderLeaf{},
		},
		{
			route: "/{paths: **}",
			style: matchStyleAll,
			want: &matchAllLeaf{
				bind: "paths",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			route, err := parser.Parse(test.route)
			assert.Nil(t, err)
			assert.Len(t, route.Segments, 1)

			segment := route.Segments[0]
			got, err := newLeaf(nil, segment, nil)
			assert.Nil(t, err)

			switch test.style {
			case matchStyleStatic:
				test.want.(*staticLeaf).segment = segment
			case matchStylePlaceholder:
				test.want.(*placeholderLeaf).segment = segment
			case matchStyleAll:
				test.want.(*matchAllLeaf).segment = segment
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
			got, err := newLeaf(nil, segment, nil)
			assert.Nil(t, err)

			leaf := got.(*regexLeaf)
			assert.Equal(t, test.wantRegexp, leaf.regexp.String())
			assert.Equal(t, test.wantBinds, leaf.binds)
		})
	}
}
