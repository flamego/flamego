// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	strptr := func(s string) *string {
		return &s
	}

	t.Run("valid routes", func(t *testing.T) {
		tests := []struct {
			route string
			want  *Route
		}{
			{
				route: "/webapi",
				want: &Route{
					Segments: []Segment{
						{
							Elements: []SegmentElement{
								{Ident: strptr("webapi")},
							},
						},
					},
				},
			},
			{
				route: "/webapi/users",
				want: &Route{
					Segments: []Segment{
						{
							Elements: []SegmentElement{
								{Ident: strptr("webapi")},
							},
						}, {
							Elements: []SegmentElement{
								{Ident: strptr("users")},
							},
						},
					},
				},
			},
			{
				route: "/{name}",
				want: &Route{
					Segments: []Segment{
						{
							Elements: []SegmentElement{
								{BindIdent: strptr("name")},
							},
						},
					},
				},
			},
			{
				route: "/webapi/{name-1}/{name-2: [a-z0-9]{7, 40}}",
				want: &Route{
					Segments: []Segment{
						{
							Elements: []SegmentElement{
								{Ident: strptr("webapi")},
							},
						}, {
							Elements: []SegmentElement{
								{BindIdent: strptr("name-1")},
							},
						}, {
							Elements: []SegmentElement{
								{
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "name-2",
												Value: &BindParameterValue{Regex: strptr("[a-z0-9]{7, 40}")},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// todo
		}
		for _, test := range tests {
			t.Run(test.route, func(t *testing.T) {
				got, err := parser.Parse(test.route)
				assert.Nil(t, err)
				assert.Equal(t, test.want, got)
			})
		}
	})

	t.Run("invalid routes", func(t *testing.T) {
		// todo
	})
}
