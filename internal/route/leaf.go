// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// MatchStyle is the match style of a tree or leaf.
type MatchStyle int8

// NOTE: The order of types matters, which determines the matching priority.
const (
	matchStyleStatic      MatchStyle = iota // e.g. /webapi
	matchStyleRegex                         // e.g. /webapi/{name: /[0-9]+/}
	matchStylePlaceholder                   // e.g. /webapi/{name}
	matchStyleAll                           // e.g. /webapi/{name: **}
)

// Leaf is a leaf derived from a segment.
type Leaf interface {
	// Parent returns the parent tree the leaf belongs to.
	Parent() *Tree
	// Optional returns true if the leaf is optional.
	Optional() bool
	// MatchStyle returns the match style of the leaf.
	MatchStyle() MatchStyle
	// URLPath fills in bind parameters with given values to build the "path"
	// portion of the URL. If withOptional is true, the path will include the
	// current leaf when it is optional; otherwise, the current leaf is excluded.
	URLPath(vals map[string]string, withOptional bool) string
}

// baseLeaf contains common fields for any leaf.
type baseLeaf struct {
	parent  *Tree    // The parent tree this leaf belongs to.
	route   *Route   // The route that the segment belongs to.
	segment *Segment // The segment that the leaf is derived from.
	handler Handler  // The handler bound to the leaf.
}

func (l *baseLeaf) Parent() *Tree {
	return l.parent
}

func (l *baseLeaf) Optional() bool {
	return l.segment.Optional
}

func (l *baseLeaf) URLPath(vals map[string]string, withOptional bool) string {
	var buf bytes.Buffer
	for _, s := range l.route.Segments {
		if s.Optional && !withOptional {
			break
		}

		buf.WriteString("/")
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
			buf.WriteString(e.BindParameters.Parameters[0].Ident)
			buf.WriteString("}")
		}
		s.str = buf.String()
	}

	pairs := make([]string, 0, len(vals)*2)
	for k, v := range vals {
		pairs = append(pairs, "{"+k+"}", v)
	}
	return strings.NewReplacer(pairs...).Replace(buf.String())
}

// staticLeaf is a leaf with a static match style.
type staticLeaf struct {
	baseLeaf
}

func (l *staticLeaf) MatchStyle() MatchStyle {
	return matchStyleStatic
}

// regexLeaf is a leaf with a regex match style.
type regexLeaf struct {
	baseLeaf
	regexp *regexp.Regexp // The regexp for the leaf.
	binds  []string       // The list of bind parameters.
}

func (l *regexLeaf) MatchStyle() MatchStyle {
	return matchStyleRegex
}

// placeholderLeaf is a leaf with a placeholder match style.
type placeholderLeaf struct {
	baseLeaf
}

func (l *placeholderLeaf) MatchStyle() MatchStyle {
	return matchStylePlaceholder
}

// placeholderLeaf is a leaf with a match all style.
type matchAllLeaf struct {
	baseLeaf
	bind string // The name of the bind parameter.
}

func (l *matchAllLeaf) MatchStyle() MatchStyle {
	return matchStyleAll
}

// newLeaf creates and returns a new Leaf derived from the given segment.
func newLeaf(t *Tree, r *Route, s *Segment, h Handler) (Leaf, error) {
	if len(s.Elements) == 0 {
		return nil, errors.Errorf("empty segment in position %d", s.Pos.Offset)
	}

	// Fast path for static route
	if len(s.Elements) == 1 && s.Elements[0].Ident != nil {
		return &staticLeaf{
			baseLeaf: baseLeaf{
				parent:  t,
				route:   r,
				segment: s,
				handler: h,
			},
		}, nil
	}

	// Fast path for placeholder
	if len(s.Elements) == 1 && s.Elements[0].BindIdent != nil {
		return &placeholderLeaf{
			baseLeaf: baseLeaf{
				parent:  t,
				route:   r,
				segment: s,
				handler: h,
			},
		}, nil
	}

	// Fast path for match all
	if len(s.Elements) == 1 &&
		s.Elements[0].BindParameters != nil &&
		len(s.Elements[0].BindParameters.Parameters) == 1 &&
		s.Elements[0].BindParameters.Parameters[0].Value.Literal != nil &&
		*s.Elements[0].BindParameters.Parameters[0].Value.Literal == "**" {
		return &matchAllLeaf{
			baseLeaf: baseLeaf{
				parent:  t,
				route:   r,
				segment: s,
				handler: h,
			},
			bind: s.Elements[0].BindParameters.Parameters[0].Ident,
		}, nil
	}

	// The only remaining style is regex
	binds := make([]string, 0, len(s.Elements))
	buf := bytes.NewBufferString("^")
	for _, e := range s.Elements {
		if e.Ident != nil {
			// Dots (".") may appear as literals, we need to escape them in a regex.
			buf.WriteString(strings.ReplaceAll(*e.Ident, ".", `\.`))
			continue
		} else if e.BindIdent != nil {
			binds = append(binds, *e.BindIdent)
			buf.WriteString("(.+)")
			continue
		} else if e.BindParameters == nil || len(e.BindParameters.Parameters) == 0 {
			return nil, errors.Errorf("empty segment element in position %d", e.Pos.Offset)
		}

		for _, p := range e.BindParameters.Parameters {
			if p.Value.Regex == nil {
				return nil, errors.Errorf("segment has non-regex literal in position %d", e.Pos.Offset)
			}

			binds = append(binds, p.Ident)
			buf.WriteString("(")
			buf.WriteString(*p.Value.Regex)
			buf.WriteString(")")
		}
	}
	buf.WriteString("$")

	re, err := regexp.Compile(buf.String())
	if err != nil {
		return nil, errors.Wrapf(err, "compile regexp near position %d", s.Pos.Offset)
	}
	return &regexLeaf{
		baseLeaf: baseLeaf{
			parent:  t,
			route:   r,
			segment: s,
			handler: h,
		},
		regexp: re,
		binds:  binds,
	}, nil
}
