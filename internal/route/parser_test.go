// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func strptr(s string) *string {
	return &s
}

func TestParser(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	t.Run("valid routes", func(t *testing.T) {
		tests := []struct {
			route string
			want  *Route
		}{
			{
				route: "/webapi",
				want: &Route{
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									Ident: strptr("webapi"),
								},
							},
						},
					},
				},
			},
			{
				route: "/webapi/users",
				want: &Route{
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									Ident: strptr("webapi"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 7,
								Line:   1,
								Column: 8,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 8,
										Line:   1,
										Column: 9,
									},
									EndPos: lexer.Position{
										Offset: 13,
										Line:   1,
										Column: 14,
									},
									Ident: strptr("users"),
								},
							},
						},
					},
				},
			},
			{
				route: "/webapi/users/?{id}",
				want: &Route{
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									Ident: strptr("webapi"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 7,
								Line:   1,
								Column: 8,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 8,
										Line:   1,
										Column: 9,
									},
									EndPos: lexer.Position{
										Offset: 13,
										Line:   1,
										Column: 14,
									},
									Ident: strptr("users"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 13,
								Line:   1,
								Column: 14,
							},
							Optional: true,
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 15,
										Line:   1,
										Column: 16,
									},
									EndPos: lexer.Position{
										Offset: 19,
										Line:   1,
										Column: 20,
									},
									BindIdent: strptr("id"),
								},
							},
						},
					},
				},
			},
			{
				route: "/{name}",
				want: &Route{
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									BindIdent: strptr("name"),
								},
							},
						},
					},
				},
			},
			{
				route: "/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}",
				want: &Route{
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									Ident: strptr("webapi"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 7,
								Line:   1,
								Column: 8,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 8,
										Line:   1,
										Column: 9,
									},
									EndPos: lexer.Position{
										Offset: 16,
										Line:   1,
										Column: 17,
									},
									BindIdent: strptr("name-1"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 16,
								Line:   1,
								Column: 17,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 17,
										Line:   1,
										Column: 18,
									},
									EndPos: lexer.Position{
										Offset: 44,
										Line:   1,
										Column: 45,
									},
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "name-2",
												Value: BindParameterValue{Regex: strptr("[a-z0-9]{7, 40}")},
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
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									Ident: strptr("webapi"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 7,
								Line:   1,
								Column: 8,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 8,
										Line:   1,
										Column: 9,
									},
									EndPos: lexer.Position{
										Offset: 16,
										Line:   1,
										Column: 17,
									},
									BindIdent: strptr("name-1"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 16,
								Line:   1,
								Column: 17,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 17,
										Line:   1,
										Column: 18,
									},
									EndPos: lexer.Position{
										Offset: 44,
										Line:   1,
										Column: 45,
									},
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "name-2",
												Value: BindParameterValue{Regex: strptr("[a-z0-9]{7, 40}")},
											},
										},
									},
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 44,
								Line:   1,
								Column: 45,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 45,
										Line:   1,
										Column: 46,
									},
									EndPos: lexer.Position{
										Offset: 59,
										Line:   1,
										Column: 60,
									},
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "year",
												Value: BindParameterValue{Literal: strptr("regex2")},
											},
										},
									},
								},
								{
									Pos: lexer.Position{
										Offset: 59,
										Line:   1,
										Column: 60,
									},
									EndPos: lexer.Position{
										Offset: 60,
										Line:   1,
										Column: 61,
									},
									Ident: strptr("-"),
								},
								{
									Pos: lexer.Position{
										Offset: 60,
										Line:   1,
										Column: 61,
									},
									EndPos: lexer.Position{
										Offset: 71,
										Line:   1,
										Column: 72,
									},
									BindIdent: strptr("month-day"),
								},
							},
						},
					},
				},
			},
			{
				// NOTE: Extra spaces before "3" is on purpose to test consecutive spaces.
				route: "/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}/{year: regex2}-{month-day}/{**: **, capture:  3}",
				want: &Route{
					Segments: []*Segment{
						{
							Pos: lexer.Position{
								Offset: 0,
								Line:   1,
								Column: 1,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 1,
										Line:   1,
										Column: 2,
									},
									EndPos: lexer.Position{
										Offset: 7,
										Line:   1,
										Column: 8,
									},
									Ident: strptr("webapi"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 7,
								Line:   1,
								Column: 8,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 8,
										Line:   1,
										Column: 9,
									},
									EndPos: lexer.Position{
										Offset: 16,
										Line:   1,
										Column: 17,
									},
									BindIdent: strptr("name-1"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 16,
								Line:   1,
								Column: 17,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 17,
										Line:   1,
										Column: 18,
									},
									EndPos: lexer.Position{
										Offset: 44,
										Line:   1,
										Column: 45,
									},
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "name-2",
												Value: BindParameterValue{Regex: strptr("[a-z0-9]{7, 40}")},
											},
										},
									},
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 44,
								Line:   1,
								Column: 45,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 45,
										Line:   1,
										Column: 46,
									},
									EndPos: lexer.Position{
										Offset: 59,
										Line:   1,
										Column: 60,
									},
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "year",
												Value: BindParameterValue{Literal: strptr("regex2")},
											},
										},
									},
								},
								{
									Pos: lexer.Position{
										Offset: 59,
										Line:   1,
										Column: 60,
									},
									EndPos: lexer.Position{
										Offset: 60,
										Line:   1,
										Column: 61,
									},
									Ident: strptr("-"),
								},
								{
									Pos: lexer.Position{
										Offset: 60,
										Line:   1,
										Column: 61,
									},
									EndPos: lexer.Position{
										Offset: 71,
										Line:   1,
										Column: 72,
									},
									BindIdent: strptr("month-day"),
								},
							},
						}, {
							Pos: lexer.Position{
								Offset: 71,
								Line:   1,
								Column: 72,
							},
							Elements: []SegmentElement{
								{
									Pos: lexer.Position{
										Offset: 72,
										Line:   1,
										Column: 73,
									},
									EndPos: lexer.Position{
										Offset: 93,
										Line:   1,
										Column: 94,
									},
									BindParameters: &BindParameters{
										Parameters: []BindParameter{
											{
												Ident: "**",
												Value: BindParameterValue{Literal: strptr("**")},
											}, {
												Ident: "capture",
												Value: BindParameterValue{Literal: strptr("3")},
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
