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
	matchStyleNone        MatchStyle = iota
	matchStyleStatic                 // e.g. /webapi
	matchStyleRegex                  // e.g. /webapi/{name: /[0-9]+/}
	matchStylePlaceholder            // e.g. /webapi/{name}
	matchStyleAll                    // e.g. /webapi/{name: **}
)

// Leaf is a leaf derived from a segment.
type Leaf interface {
	// URLPath fills in bind parameters with given values to build the "path"
	// portion of the URL. If withOptional is true, the path will include the
	// current leaf when it is optional; otherwise, the current leaf is excluded.
	URLPath(vals map[string]string, withOptional bool) string

	// getParent returns the parent tree the leaf belongs to.
	getParent() Tree
	// getSegment returns the segment that the leaf is derived from.
	getSegment() *Segment
	// getMatchStyle returns the match style of the leaf.
	getMatchStyle() MatchStyle
	// todo
	match(segment string, params Params) bool
}

// baseLeaf contains common fields for any leaf.
type baseLeaf struct {
	parent  Tree     // The parent tree this leaf belongs to.
	route   *Route   // The route that the segment belongs to.
	segment *Segment // The segment that the leaf is derived from.
	handler Handler  // The handler bound to the leaf.
}

func (l *baseLeaf) getParent() Tree {
	return l.parent
}

func (l *baseLeaf) getSegment() *Segment {
	return l.segment
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

func (l *staticLeaf) getMatchStyle() MatchStyle {
	return matchStyleStatic
}

func (l *staticLeaf) match(segment string, _ Params) bool {
	return l.segment.String()[1:] == segment // Skip the leading "/"
}

// regexLeaf is a leaf with a regex match style.
type regexLeaf struct {
	baseLeaf
	regexp *regexp.Regexp // The regexp for the leaf.
	binds  []string       // The list of bind parameters.
}

func (l *regexLeaf) getMatchStyle() MatchStyle {
	return matchStyleRegex
}

func (l *regexLeaf) match(segment string, params Params) bool {
	submatches := l.regexp.FindStringSubmatch(segment)
	if len(submatches) != len(l.binds)+1 {
		return false
	}

	for i, bind := range l.binds {
		params[bind] = submatches[i+1]
	}
	return true
}

// placeholderLeaf is a leaf with a placeholder match style.
type placeholderLeaf struct {
	baseLeaf
	bind string // The name of the bind parameter.
}

func (l *placeholderLeaf) getMatchStyle() MatchStyle {
	return matchStylePlaceholder
}

func (l *placeholderLeaf) match(segment string, params Params) bool {
	params[l.bind] = segment
	return true
}

// placeholderLeaf is a leaf with a match all style.
type matchAllLeaf struct {
	baseLeaf
	bind string // The name of the bind parameter.
}

func (l *matchAllLeaf) getMatchStyle() MatchStyle {
	return matchStyleAll
}

func (l *matchAllLeaf) match(segment string, params Params) bool {
	params[l.bind] = segment
	return true
}

// isMatchStyleStatic returns true if the Segment is static match style.
func isMatchStyleStatic(s *Segment) bool {
	return len(s.Elements) == 1 && s.Elements[0].Ident != nil
}

// checkMatchStylePlaceholder returns true if the Segment is placeholder match
// style, along with // its bind parameter name.
func checkMatchStylePlaceholder(s *Segment) (bind string, ok bool) {
	if len(s.Elements) == 1 &&
		s.Elements[0].BindIdent != nil {
		return *s.Elements[0].BindIdent, true
	}
	return "", false
}

// checkMatchStyleAll returns true if the Segment is match all style, along with
// its bind parameter name.
func checkMatchStyleAll(s *Segment) (bind string, ok bool) {
	if len(s.Elements) == 1 &&
		s.Elements[0].BindParameters != nil &&
		len(s.Elements[0].BindParameters.Parameters) == 1 &&
		s.Elements[0].BindParameters.Parameters[0].Value.Literal != nil &&
		*s.Elements[0].BindParameters.Parameters[0].Value.Literal == "**" {
		return s.Elements[0].BindParameters.Parameters[0].Ident, true
	}
	return "", false
}

// constructMatchStyleRegex constructs a regexp from the Segment (having the
// assumption that it's regex match style), along with bind parameter names in
// the same order as regexp's sub-matches.
func constructMatchStyleRegex(s *Segment) (*regexp.Regexp, []string, error) {
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
			return nil, nil, errors.Errorf("empty segment element in position %d", e.Pos.Offset)
		}

		for _, p := range e.BindParameters.Parameters {
			if p.Value.Regex == nil {
				return nil, nil, errors.Errorf("segment has non-regex literal in position %d", e.Pos.Offset)
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
		return nil, nil, errors.Wrapf(err, "compile regexp near position %d", s.Pos.Offset)
	}
	return re, binds, nil
}

// newLeaf creates and returns a new Leaf derived from the given segment.
func newLeaf(parent Tree, r *Route, s *Segment, h Handler) (Leaf, error) {
	if len(s.Elements) == 0 {
		return nil, errors.Errorf("empty segment in position %d", s.Pos.Offset)
	}

	if isMatchStyleStatic(s) {
		return &staticLeaf{
			baseLeaf: baseLeaf{
				parent:  parent,
				route:   r,
				segment: s,
				handler: h,
			},
		}, nil
	}

	if bind, ok := checkMatchStylePlaceholder(s); ok {
		return &placeholderLeaf{
			baseLeaf: baseLeaf{
				parent:  parent,
				route:   r,
				segment: s,
				handler: h,
			},
			bind: bind,
		}, nil
	}

	if bind, ok := checkMatchStyleAll(s); ok {
		return &matchAllLeaf{
			baseLeaf: baseLeaf{
				parent:  parent,
				route:   r,
				segment: s,
				handler: h,
			},
			bind: bind,
		}, nil
	}

	// The only remaining style is regex.
	re, binds, err := constructMatchStyleRegex(s)
	if err != nil {
		return nil, err
	}
	return &regexLeaf{
		baseLeaf: baseLeaf{
			parent:  parent,
			route:   r,
			segment: s,
			handler: h,
		},
		regexp: re,
		binds:  binds,
	}, nil
}
