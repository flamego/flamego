// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"sync"

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
	Ident string             `parser:"@Ident ':' ' '*"`
	Value BindParameterValue `parser:"@@"`
}

// BindParameters is a set of bind parameters that are separated by commas
// (",").
type BindParameters struct {
	Parameters []BindParameter `parser:"( @@ ( ',' ' '* @@ )* )+"`
}

// SegmentElement is a single segment element containing either identifier, bind
// identifier or a list of bind parameters. Bind identifier and the list of bind
// parameters are surrounded by brackets ("{}").
type SegmentElement struct {
	Pos            lexer.Position
	EndPos         lexer.Position
	Ident          *string         `parser:"  @Ident"`
	BindIdent      *string         `parser:"| '{' @Ident '}'"`
	BindParameters *BindParameters `parser:"| '{' @@ '}'"`
}

// Segment is a single segment containing multiple elements.
type Segment struct {
	Pos      lexer.Position
	Slash    string           `parser:"'/'"`
	Optional bool             `parser:"@'?'?"`
	Elements []SegmentElement `parser:"@@*"`

	strOnce sync.Once `parser:"-"`
	str     string    `parser:"-"`
}

// String returns the string representation of the Segment, which basically
// traverses the Segment AST to reconstruct the original string.
func (s *Segment) String() string {
	s.strOnce.Do(func() {
		var buf bytes.Buffer
		buf.WriteString("/")

		if s.Optional {
			buf.WriteString("?")
		}

		for _, e := range s.Elements {
			if e.Ident != nil {
				buf.WriteString(*e.Ident)
				continue
			} else if e.BindIdent != nil {
				buf.WriteString("{")
				buf.WriteString(*e.BindIdent)
				buf.WriteString("}")
				continue
			} else if e.BindParameters == nil || len(e.BindParameters.Parameters) == 0 {
				// This element is empty, which should never happen, but just in case
				buf.WriteString("???")
				continue
			}

			buf.WriteString("{")
			for i, p := range e.BindParameters.Parameters {
				buf.WriteString(p.Ident)
				buf.WriteString(": ")

				switch {
				case p.Value.Literal != nil:
					buf.WriteString(*p.Value.Literal)
				case p.Value.Regex != nil:
					buf.WriteString("/")
					buf.WriteString(*p.Value.Regex)
					buf.WriteString("/")
				default:
					// This parameter has no value, which should never happen, but just in case
					buf.WriteString("???")
				}

				if len(e.BindParameters.Parameters) > i+1 {
					buf.WriteString(", ")
				}
			}
			buf.WriteString("}")
		}
		s.str = buf.String()
	})
	return s.str
}

// Route is a single route containing segments that are separated by slashes
// ("/").
type Route struct {
	Segments []*Segment `parser:"@@+"`

	strOnce sync.Once `parser:"-"`
	str     string    `parser:"-"`
}

// String returns the string representation of the Route, which basically
// traverses the Route AST to reconstruct the original route string.
func (r *Route) String() string {
	r.strOnce.Do(func() {
		var buf bytes.Buffer
		for _, s := range r.Segments {
			buf.WriteString(s.String())
		}
		r.str = buf.String()
	})
	return r.str
}
