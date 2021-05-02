// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
	"github.com/pkg/errors"
)

// todo
type Parser struct {
	parser *participle.Parser
}

// todo
func (p *Parser) Parse(s string) (*Route, error) {
	ast := &Route{}
	return ast, p.parser.ParseString("", s, ast)
}

// todo
func NewParser() (*Parser, error) {
	lexer, err := stateful.New(
		stateful.Rules{
			"Root": {
				{"Segment", `/`, stateful.Push("Segment")},
			},
			"Segment": {
				stateful.Include("Common"),
				{"Bind", `{`, stateful.Push("Bind")},
				{"Segment", `/`, stateful.Push("Segment")},
			},
			"Bind": {
				stateful.Include("Common"),
				{"BindParameter", `:`, stateful.Push("BindParameter")},
				{"Bind", `{`, stateful.Push("Bind")},
				{"BindEnd", `}`, stateful.Pop()},
				{"Segment", `/`, stateful.Push("Segment")},
			},
			"BindParameter": {
				stateful.Include("Common"),
				{"BindParameterValue", `/`, stateful.Push("BindParameterValue")},
				{"BindParameterEnd", `[},]`, stateful.Pop()},
			},
			"BindParameterValue": {
				stateful.Include("Common"),
				{"Regex", `[a-zA-Z0-9*\-+.,?()\[\]{} \\]+`, nil},
				{"RegexEnd", `/`, stateful.Pop()},
			},
			"Common": {
				{"Ident", `[a-zA-Z0-9*\-]+`, nil},
				{"Whitespace", `\s`, nil},
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "new lexer")
	}

	parser, err := participle.Build(
		&Route{},
		participle.Lexer(lexer),
		participle.UseLookahead(2),
	)
	if err != nil {
		return nil, errors.Wrap(err, "build parser")
	}

	return &Parser{parser: parser}, nil
}
