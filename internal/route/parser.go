// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pkg/errors"
)

// Parser is a BNF-based route syntax parser using stateful lexer.
type Parser struct {
	parser *participle.Parser[Route]
}

// Parse parses and returns a single route.
func (p *Parser) Parse(s string) (*Route, error) {
	return p.parser.ParseString("", s)
}

// NewParser creates and returns a new Parser.
func NewParser() (*Parser, error) {
	l, err := lexer.New(
		lexer.Rules{
			"Root": {
				{Name: "Segment", Pattern: `/`, Action: lexer.Push("Segment")},
			},
			"Segment": {
				lexer.Include("Common"),
				{Name: "Optional", Pattern: `[?]`},
				{Name: "Bind", Pattern: `{`, Action: lexer.Push("Bind")},
				{Name: "Segment", Pattern: `/`, Action: lexer.Push("Segment")},
			},
			"Bind": {
				lexer.Include("Common"),
				{Name: "BindParameter", Pattern: `:`, Action: lexer.Push("BindParameter")},
				{Name: "Bind", Pattern: `{`, Action: lexer.Push("Bind")},
				{Name: "BindEnd", Pattern: `}`, Action: lexer.Pop()},
				{Name: "Segment", Pattern: `/`, Action: lexer.Push("Segment")},
			},
			"BindParameter": {
				lexer.Include("Common"),
				{Name: "BindParameterRegexValue", Pattern: `/`, Action: lexer.Push("BindParameterRegexValue")},
				{Name: "BindParameterEnd", Pattern: `[},]`, Action: lexer.Pop()},
			},
			"BindParameterRegexValue": {
				{Name: "Regex", Pattern: `[a-zA-Z0-9*\-+._,?()\[\]{} \\\|]+`},
				{Name: "RegexEnd", Pattern: `/`, Action: lexer.Pop()},
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

	parser, err := participle.Build[Route](
		participle.Lexer(l),
		participle.UseLookahead(2),
	)
	if err != nil {
		return nil, errors.Wrap(err, "build parser")
	}

	return &Parser{parser: parser}, nil
}
