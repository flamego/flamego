// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// MatchStyle is the match style of a tree or leaf.
type MatchStyle int8

// NOTE: The order of types matters, which determines the matching priority.
const (
	matchStyleNone        MatchStyle = iota
	matchStyleStatic                 // e.g. "/webapi"
	matchStyleRegex                  // e.g. "/webapi/{name: /[0-9]+/}"
	matchStylePlaceholder            // e.g. "/webapi/{name}"
	matchStyleAll                    // e.g. "/webapi/{name: **}", "/webapi/{**}"
)

// Leaf is a leaf derived from a segment.
type Leaf interface {
	// SetHeaderMatcher sets the HeaderMatcher for the leaf.
	SetHeaderMatcher(m *HeaderMatcher)

	// URLPath fills in bind parameters with given values to build the "path"
	// portion of the URL. If `withOptional` is true, the path will include the
	// current leaf when it is optional; otherwise, the current leaf is excluded.
	URLPath(vals map[string]string, withOptional bool) string
	// Route returns the string representation of the original route.
	Route() string
	// Handler the Handler that is associated with the leaf.
	Handler() Handler
	// Static returns true if the leaf and all ancestors are static routes.
	Static() bool

	// getParent returns the parent tree the leaf belongs to.
	getParent() Tree
	// getSegment returns the segment that the leaf is derived from.
	getSegment() *Segment
	// getMatchStyle returns the match style of the leaf.
	getMatchStyle() MatchStyle
	// match returns true if the leaf matches the segment, values of bind parameters
	// are stored in the `Params`.
	match(segment string, params Params, header http.Header) bool
}

// baseLeaf contains common fields for any leaf.
type baseLeaf struct {
	parent        Tree           // The parent tree this leaf belongs to.
	route         *Route         // The route that the segment belongs to.
	segment       *Segment       // The segment that the leaf is derived from.
	handler       Handler        // The handler bound to the leaf.
	headerMatcher *HeaderMatcher // The matcher for header values.
}

func (l *baseLeaf) getParent() Tree {
	return l.parent
}

func (l *baseLeaf) getSegment() *Segment {
	return l.segment
}

func (l *baseLeaf) SetHeaderMatcher(m *HeaderMatcher) {
	l.headerMatcher = m
}

func (l *baseLeaf) matchHeader(header http.Header) bool {
	return l.headerMatcher == nil || l.headerMatcher.Match(header)
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

func (l *baseLeaf) Route() string {
	return l.route.String()
}

func (l *baseLeaf) Handler() Handler {
	return l.handler
}

func (*baseLeaf) Static() bool {
	return false
}

// staticLeaf is a leaf with a static match style.
type staticLeaf struct {
	baseLeaf
	literals string
}

func (*staticLeaf) getMatchStyle() MatchStyle {
	return matchStyleStatic
}

func (l *staticLeaf) match(segment string, _ Params, header http.Header) bool {
	return l.literals == segment && l.matchHeader(header)
}

func (l *staticLeaf) Static() bool {
	ancestor := l.parent
	for ancestor != nil {
		if ancestor.getMatchStyle() > matchStyleStatic {
			return false
		}
		ancestor = ancestor.getParent()
	}
	return true
}

// regexLeaf is a leaf with a regex match style.
type regexLeaf struct {
	baseLeaf
	regexp *regexp.Regexp // The regexp for the leaf.
	binds  []string       // The list of bind parameters.
}

func (*regexLeaf) getMatchStyle() MatchStyle {
	return matchStyleRegex
}

func (l *regexLeaf) match(segment string, params Params, header http.Header) bool {
	submatches := l.regexp.FindStringSubmatch(segment)
	if len(submatches) < len(l.binds)+1 {
		return false
	}

	if !l.matchHeader(header) {
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

func (*placeholderLeaf) getMatchStyle() MatchStyle {
	return matchStylePlaceholder
}

func (l *placeholderLeaf) match(segment string, params Params, header http.Header) bool {
	if !l.matchHeader(header) {
		return false
	}
	params[l.bind] = segment
	return true
}

// placeholderLeaf is a leaf with a match all style.
type matchAllLeaf struct {
	baseLeaf
	bind    string // The name of the bind parameter.
	capture int    // The capture limit of the bind parameter. Non-positive means unlimited.
}

func (*matchAllLeaf) getMatchStyle() MatchStyle {
	return matchStyleAll
}

func (l *matchAllLeaf) match(segment string, params Params, header http.Header) bool {
	if !l.matchHeader(header) {
		return false
	}
	params[l.bind] = segment
	return true
}

// matchAll matches all remaining segments up to the capture limit (when
// defined). The `path` should be original request path, `segment` should NOT be
// unescaped by the caller. It returns true if segments are captured within the
// limit, and the capture result is stored in `params`.
func (l *matchAllLeaf) matchAll(path, segment string, next int, params Params, header http.Header) bool {
	// Do `next-1` because "next" starts at the next character of preceding "/"; do
	// `strings.Count()+1` because the segment itself also counts. E.g. "webapi" +
	// "users/events" => 3
	if l.capture > 0 && l.capture < strings.Count(path[next-1:], "/")+1 {
		return false
	}
	if !l.matchHeader(header) {
		return false
	}

	params[l.bind] = segment + "/" + path[next:]
	return true
}

// isMatchStyleStatic returns true if the Segment is static match style.
func isMatchStyleStatic(s *Segment) bool {
	return len(s.Elements) == 1 && s.Elements[0].Ident != nil
}

// checkMatchStylePlaceholder returns true if the Segment is placeholder match
// style, along with its bind parameter name. The BindIdent "**" is ignored as a
// special case of matchStyleAll.
func checkMatchStylePlaceholder(s *Segment) (bind string, ok bool) {
	if len(s.Elements) == 1 &&
		s.Elements[0].BindIdent != nil &&
		*s.Elements[0].BindIdent != "**" {
		return *s.Elements[0].BindIdent, true
	}
	return "", false
}

// checkMatchStyleAll returns true if the Segment is match all style, along with
// its bind parameter name and capture limit. The capture is 0 when undefined.
// The BindIdent "**" is treated as a special case for "{**: **}".
func checkMatchStyleAll(s *Segment) (bind string, capture int, ok bool) {
	// Special case for "{**}"
	if len(s.Elements) == 1 &&
		s.Elements[0].BindIdent != nil &&
		*s.Elements[0].BindIdent == "**" &&
		s.Elements[0].BindParameters == nil {
		return "**", 0, true
	}

	// Check for "{<BindIdent>: **}"
	if len(s.Elements) == 0 ||
		s.Elements[0].BindParameters == nil ||
		len(s.Elements[0].BindParameters.Parameters) == 0 ||
		s.Elements[0].BindParameters.Parameters[0].Value.Literal == nil ||
		*s.Elements[0].BindParameters.Parameters[0].Value.Literal != "**" {
		return "", 0, false
	}

	bind = s.Elements[0].BindParameters.Parameters[0].Ident

	if len(s.Elements[0].BindParameters.Parameters) > 1 &&
		s.Elements[0].BindParameters.Parameters[1].Ident == "capture" &&
		s.Elements[0].BindParameters.Parameters[1].Value.Literal != nil {
		capture, _ = strconv.Atoi(*s.Elements[0].BindParameters.Parameters[1].Value.Literal)
	}

	return bind, capture, true
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

// getParentBindSet returns a set of all bind parameters defined in parent
// trees.
func getParentBindSet(parent Tree) map[string]struct{} {
	bindSet := make(map[string]struct{})
	ancestor := parent
	for ancestor != nil {
		for _, bind := range ancestor.getBinds() {
			bindSet[bind] = struct{}{}
		}
		ancestor = ancestor.getParent()
	}
	return bindSet
}

// newLeaf creates and returns a new Leaf derived from the given segment.
func newLeaf(parent Tree, r *Route, s *Segment, h Handler) (Leaf, error) {
	// Based on the syntax definition, the only possible case to have a leaf with no
	// segment elements is when the entire route is "/". Therefore, we can safely
	// treat it as a static match with an empty string.
	if len(s.Elements) == 0 || isMatchStyleStatic(s) {
		return &staticLeaf{
			baseLeaf: baseLeaf{
				parent:  parent,
				route:   r,
				segment: s,
				handler: h,
			},
			literals: strings.TrimLeft(s.String(), "/?"),
		}, nil
	}

	parentBindSet := getParentBindSet(parent)

	if bind, ok := checkMatchStylePlaceholder(s); ok {
		if _, exists := parentBindSet[bind]; exists {
			return nil, errors.Errorf("duplicated bind parameter %q in position %d", bind, s.Pos.Offset)
		}
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

	if bind, capture, ok := checkMatchStyleAll(s); ok {
		if _, exists := parentBindSet[bind]; exists {
			return nil, errors.Errorf("duplicated bind parameter %q in position %d", bind, s.Pos.Offset)
		}
		return &matchAllLeaf{
			baseLeaf: baseLeaf{
				parent:  parent,
				route:   r,
				segment: s,
				handler: h,
			},
			bind:    bind,
			capture: capture,
		}, nil
	}

	// The only remaining style is regex.
	re, binds, err := constructMatchStyleRegex(s)
	if err != nil {
		return nil, err
	}

	for _, bind := range binds {
		if _, exists := parentBindSet[bind]; exists {
			return nil, errors.Errorf("duplicated bind parameter %q in position %d", bind, s.Pos.Offset)
		}
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
