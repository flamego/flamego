// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
	"github.com/pkg/errors"
)

// Parser is a BNF-based route syntax parser using stateful lexer.
type Parser struct {
	parser *participle.Parser
}

// Parse parses and returns a single route.
func (p *Parser) Parse(s string) (*Route, error) {
	ast := &Route{}
	return ast, p.parser.ParseString("", s, ast)
}

// NewParser creates and returns a new Parser.
func NewParser() (*Parser, error) {
	lexer, err := stateful.New(
		stateful.Rules{
			"Root": {
				{Name: "Segment", Pattern: `/`, Action: stateful.Push("Segment")},
			},
			"Segment": {
				stateful.Include("Common"),
				{Name: "Optional", Pattern: `[?]`},
				{Name: "Bind", Pattern: `{`, Action: stateful.Push("Bind")},
				{Name: "Segment", Pattern: `/`, Action: stateful.Push("Segment")},
			},
			"Bind": {
				stateful.Include("Common"),
				{Name: "BindParameter", Pattern: `:`, Action: stateful.Push("BindParameter")},
				{Name: "Bind", Pattern: `{`, Action: stateful.Push("Bind")},
				{Name: "BindEnd", Pattern: `}`, Action: stateful.Pop()},
				{Name: "Segment", Pattern: `/`, Action: stateful.Push("Segment")},
			},
			"BindParameter": {
				stateful.Include("Common"),
				{Name: "BindParameterRegexValue", Pattern: `/`, Action: stateful.Push("BindParameterRegexValue")},
				{Name: "BindParameterEnd", Pattern: `[},]`, Action: stateful.Pop()},
			},
			"BindParameterRegexValue": {
				{Name: "Regex", Pattern: `[a-zA-Z0-9*\-+._,?()\[\]{} \\\|]+`},
				{Name: "RegexEnd", Pattern: `/`, Action: stateful.Pop()},
			},
			"Common": {
				// All legal URI characters that are defined in RFC 3986.
				// FIXME: `[]:,` are not allowed, since it may affect the binding processing
				{Name: "Ident", Pattern: `[a-zA-Z0-9\-._~@!$&'()*+;%=]+`},
				{Name: "Whitespace", Pattern: `\s`},
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
