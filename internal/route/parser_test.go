// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
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
				route: "/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}",
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
			{
				route: "/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}/{year: regex2}-{month-day}",
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
						}, {
							Elements: []SegmentElement{
								{
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "year",
												Value: &BindParameterValue{Literal: strptr("regex2")},
											},
										},
									},
								},
								{Ident: strptr("-")},
								{BindIdent: strptr("month-day")},
							},
						},
					},
				},
			},
			{
				// NOTE: Extra spaces before "3" is on purpose to test consecutive spaces.
				route: "/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}/{year: regex2}-{month-day}/{**: **, capture:  3}",
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
						}, {
							Elements: []SegmentElement{
								{
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "year",
												Value: &BindParameterValue{Literal: strptr("regex2")},
											},
										},
									},
								},
								{Ident: strptr("-")},
								{BindIdent: strptr("month-day")},
							},
						}, {
							Elements: []SegmentElement{
								{
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "**",
												Value: &BindParameterValue{Literal: strptr("**")},
											}, {
												Ident: "capture",
												Value: &BindParameterValue{Literal: strptr("3")},
											},
										},
									},
								},
							},
						},
					},
				},
			},
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
		tests := []struct {
			name    string
			route   string
			wantErr string
		}{
			{
				name:    "missing leading slash",
				route:   "webapi",
				wantErr: `1:1: invalid input text "webapi"`,
			},
			{
				name:    "missing opening bracket",
				route:   "/name}",
				wantErr: `1:6: invalid input text "}"`,
			},
			{
				name:    "missing closing bracket",
				route:   "/{name",
				wantErr: `1:7: unexpected token "<EOF>" (expected "}")`,
			},
			{
				name:    "no surroundings for regex",
				route:   "/{name: [a-z0-9]{7, 40}}",
				wantErr: `1:9: invalid input text "[a-z0-9]{7, 40}}"`,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				_, err := parser.Parse(test.route)
				got := fmt.Sprintf("%v", err)
				assert.Equal(t, test.wantErr, got)
			})
		}
	})
}
