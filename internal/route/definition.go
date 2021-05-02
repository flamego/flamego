// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

// todo
type BindParameterValue struct {
	Literal *string `parser:"  @Ident"`
	Regex   *string `parser:"| '/' @Regex '/'"`
}

// todo
type BindParameter struct {
	Ident string              `parser:"@Ident ':' ' '*"`
	Value *BindParameterValue `parser:"@@"`
}

// todo
type BindParameters struct {
	Parameters []BindParameter `parser:"( @@ ( ',' ' '* @@ )* )+"`
}

// todo
type SegmentElement struct {
	Ident          *string         `parser:"  @Ident"`
	BindIdent      *string         `parser:"| '{' @Ident '}'"`
	BindParameters *BindParameters `parser:"| '{' @@ '}'"`
}

// todo
type Segment struct {
	Elements []SegmentElement `parser:"@@*"`
}

// todo
type Route struct {
	Segments []Segment `parser:"( '/' @@ )+"`
}
