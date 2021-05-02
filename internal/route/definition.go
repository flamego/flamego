// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// BindParameterValue is a single bind parameter value containing either literal
// or regex, the latter is surrounded by slashes ("/").
type BindParameterValue struct {
	Literal *string `parser:"  @Ident"`
	Regex   *string `parser:"| '/' @Regex '/'"`
}

// BindParameter is a single pair of bind parameter containing identifier and
// value that are separated by the colon (":").
type BindParameter struct {
	Ident string              `parser:"@Ident ':' ' '*"`
	Value *BindParameterValue `parser:"@@"`
}

// BindParameters is a set of bind parameters that are separated by commas
// (",").
type BindParameters struct {
	Parameters []BindParameter `parser:"( @@ ( ',' ' '* @@ )* )+"`
}

// SegmentElement is a single segment element containing either identifier, bind
// identifier or bind parameters. Bind identifier and bind parameters are
// surrounded by brackets ("{}").
type SegmentElement struct {
	Pos            lexer.Position
	EndPos         lexer.Position
	Ident          *string         `parser:"  @Ident"`
	BindIdent      *string         `parser:"| '{' @Ident '}'"`
	BindParameters *BindParameters `parser:"| '{' @@ '}'"`
}

// Segment is a single segment containing multiple elements.
type Segment struct {
	Elements []SegmentElement `parser:"@@*"`
}

// Route is a single route containing segments that are separated by slashes
// ("/").
type Route struct {
	Segments []Segment `parser:"( '/' @@ )+"`
}
