// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Tree is a tree derived from a segment.
type Tree interface {
	// Match matches a leaf for the given request path and provided headers, values
	// of bind parameters are stored in the `Params`. The `Params` may contain extra
	// values that do not belong to the final leaf due to backtrace.
	Match(path string, header http.Header) (Leaf, Params, bool)

	// getParent returns the parent tree. The root tree does not have parent.
	getParent() Tree
	// getSegment returns the segment that the tree is derived from.
	getSegment() *Segment
	// getMatchStyle returns the match style of the tree.
	getMatchStyle() MatchStyle
	// getSubtrees returns the list of direct subtrees.
	getSubtrees() []Tree
	// getLeaves returns the list of direct leaves.
	getLeaves() []Leaf
	// setSubtrees sets the list of direct subtrees.
	setSubtrees(subtrees []Tree)
	// setLeaves sets the list of direct leaves.
	setLeaves(leaves []Leaf)
	// getBinds returns the list of bind parameters.
	getBinds() []string
	// hasMatchAllSubtree returns true if there is on match all style subtree exists.
	hasMatchAllSubtree() bool
	// hasMatchAllLeaf returns true if there is on match all style leaf exists.
	hasMatchAllLeaf() bool
	// match returns true if the tree matches the segment, values of bind parameters
	// are stored in the `Params`. The `Params` may contain extra values that do not
	// belong to the final leaf due to backtrace.
	match(segment string, params Params) bool
	// matchNextSegment advances the `next` cursor for matching next segment in the
	// request path.
	matchNextSegment(path string, next int, params Params, header http.Header) (Leaf, bool)
}

// baseTree contains common fields and methods for any tree.
type baseTree struct {
	parent   Tree     // The parent tree.
	segment  *Segment // The segment that the tree is derived from.
	subtrees []Tree   // The list of direct subtrees ordered by matching priority.
	leaves   []Leaf   // The list of direct leaves ordered by matching priority.
}

func (t *baseTree) getParent() Tree {
	return t.parent
}

func (t *baseTree) getSegment() *Segment {
	return t.segment
}

func (*baseTree) getMatchStyle() MatchStyle {
	return matchStyleNone
}

func (t *baseTree) getSubtrees() []Tree {
	return t.subtrees
}

func (t *baseTree) getLeaves() []Leaf {
	return t.leaves
}

func (t *baseTree) setSubtrees(subtrees []Tree) {
	t.subtrees = subtrees
}

func (t *baseTree) setLeaves(leaves []Leaf) {
	t.leaves = leaves
}

func (*baseTree) getBinds() []string {
	return nil
}

func (t *baseTree) hasMatchAllSubtree() bool {
	return len(t.subtrees) > 0 &&
		t.subtrees[len(t.subtrees)-1].getMatchStyle() == matchStyleAll
}

func (t *baseTree) hasMatchAllLeaf() bool {
	return len(t.leaves) > 0 &&
		t.leaves[len(t.leaves)-1].getMatchStyle() == matchStyleAll
}

func (*baseTree) match(_ string, _ Params) bool {
	panic("unreachable")
}

// Params is a set of bind parameters with their values that are extracted from
// the request path.
type Params map[string]string

// Handler is a function that can be registered to a route for handling HTTP
// requests.
type Handler func(http.ResponseWriter, *http.Request, Params)

// addLeaf adds a new leaf from the given segment.
func addLeaf(t Tree, r *Route, s *Segment, h Handler) (Leaf, error) {
	leaves := t.getLeaves()
	for _, l := range leaves {
		if l.getSegment().String() == s.String() {
			return nil, errors.Errorf("duplicated route %q", r.String())
		}
	}

	leaf, err := newLeaf(t, r, s, h)
	if err != nil {
		return nil, errors.Wrap(err, "new leaf")
	}

	// At most one match all style leaf can exist in a leaf list.
	if leaf.getMatchStyle() == matchStyleAll &&
		t.hasMatchAllLeaf() {
		return nil, errors.Errorf("duplicated match all bind parameter in position %d", s.Pos.Offset)
	}

	if leaf.getSegment().Optional {
		parent := leaf.getParent()
		if parent.getParent() != nil {
			_, err = addLeaf(parent.getParent(), r, parent.getSegment(), h)
			if err != nil {
				return nil, errors.Wrap(err, "add optional leaf to grandparent")
			}
		} else {
			_, err = addLeaf(parent, r, parent.getSegment(), h)
			if err != nil {
				return nil, errors.Wrap(err, "add optional leaf to parent")
			}
		}
	}

	// Determine leaf position by the priority of match styles.
	i := 0
	for ; i < len(leaves); i++ {
		if leaf.getMatchStyle() < leaves[i].getMatchStyle() {
			break
		}
	}

	if i == len(leaves) {
		leaves = append(leaves, leaf)
	} else {
		leaves = append(leaves[:i], append([]Leaf{leaf}, leaves[i:]...)...)
	}
	t.setLeaves(leaves)

	return leaf, nil
}

// addSubtree adds a new subtree from next segment of the route.
func addSubtree(t Tree, r *Route, next int, h Handler) (Leaf, error) {
	segment := r.Segments[next]
	for _, st := range t.getSubtrees() {
		if st.getSegment().String() == segment.String() {
			return addNextSegment(st, r, next+1, h)
		}
	}

	subtree, err := newTree(t, segment)
	if err != nil {
		return nil, errors.Wrap(err, "new tree")
	}

	// At most one match all style subtree can exist in a subtree list.
	if subtree.getMatchStyle() == matchStyleAll &&
		t.hasMatchAllSubtree() {
		return nil, errors.Errorf("duplicated match all bind parameter in position %d", segment.Pos.Offset)
	}

	// Determine subtree position by the priority of match styles.
	subtrees := t.getSubtrees()
	i := 0
	for ; i < len(subtrees); i++ {
		if subtree.getMatchStyle() < subtrees[i].getMatchStyle() {
			break
		}
	}

	if i == len(subtrees) {
		subtrees = append(subtrees, subtree)
	} else {
		subtrees = append(subtrees[:i], append([]Tree{subtree}, subtrees[i:]...)...)
	}
	t.setSubtrees(subtrees)

	return addNextSegment(subtree, r, next+1, h)
}

// addNextSegment adds next segment of the route to the tree.
func addNextSegment(t Tree, r *Route, next int, h Handler) (Leaf, error) {
	if len(r.Segments) <= next+1 {
		return addLeaf(t, r, r.Segments[next], h)
	}

	if r.Segments[next].Optional {
		return nil, errors.New("only the last segment can be optional")
	}
	return addSubtree(t, r, next, h)
}

// staticTree is a tree with a static match style.
type staticTree struct {
	baseTree
}

func (*staticTree) getMatchStyle() MatchStyle {
	return matchStyleStatic
}

func (*staticTree) getBinds() []string {
	return nil
}

func (t *staticTree) match(segment string, _ Params) bool {
	return t.segment.String()[1:] == segment // Skip the leading "/"
}

// regexTree is a tree with a regex match style.
type regexTree struct {
	baseTree
	regexp *regexp.Regexp // The regexp for the tree.
	binds  []string       // The list of bind parameters.
}

func (*regexTree) getMatchStyle() MatchStyle {
	return matchStyleRegex
}

func (t *regexTree) getBinds() []string {
	binds := make([]string, len(t.binds))
	copy(binds, t.binds)
	return binds
}

func (t *regexTree) match(segment string, params Params) bool {
	submatches := t.regexp.FindStringSubmatch(segment)
	if len(submatches) != len(t.binds)+1 {
		return false
	}

	for i, bind := range t.binds {
		params[bind] = submatches[i+1]
	}
	return true
}

// placeholderTree is a tree with a placeholder match style.
type placeholderTree struct {
	baseTree
	bind string // The name of the bind parameter.
}

func (*placeholderTree) getMatchStyle() MatchStyle {
	return matchStylePlaceholder
}

func (t *placeholderTree) getBinds() []string {
	return []string{t.bind}
}

func (t *placeholderTree) match(segment string, params Params) bool {
	params[t.bind] = segment
	return true
}

// placeholderTree is a tree with a match all style.
type matchAllTree struct {
	baseTree
	bind    string // The name of the bind parameter.
	capture int    // The capture limit of the bind parameter. Non-positive means unlimited.
}

func (*matchAllTree) getMatchStyle() MatchStyle {
	return matchStyleAll
}

func (t *matchAllTree) getBinds() []string {
	return []string{t.bind}
}

// matchAll matches all remaining segments up to the capture limit (when
// defined). The `path` should be original request path, `segment` should NOT be
// unescaped by the caller. It returns the matched leaf and true if segments are
// captured within the limit, and the capture result is stored in `params`.
func (t *matchAllTree) matchAll(path, segment string, next int, params Params, header http.Header) (Leaf, bool) {
	captured := 1 // Starts with 1 because the segment itself also count.
	for t.capture <= 0 || t.capture >= captured {
		leaf, ok := t.matchNextSegment(path, next, params, header)
		if ok {
			params[t.bind] = segment
			return leaf, true
		}

		i := strings.Index(path[next:], "/")
		if i == -1 {
			// We've reached last segment of the request path, but it should be matched by a
			// leaf not a subtree by definition.
			break
		}

		segment += "/" + path[next:next+i]
		next += i + 1
		captured++
	}
	return nil, false
}

// newTree creates and returns a new Tree derived from the given segment.
func newTree(parent Tree, s *Segment) (Tree, error) {
	if len(s.Elements) == 0 {
		return nil, errors.Errorf("empty segment in position %d", s.Pos.Offset)
	}

	if isMatchStyleStatic(s) {
		return &staticTree{
			baseTree: baseTree{
				parent:  parent,
				segment: s,
			},
		}, nil
	}

	parentBindSet := getParentBindSet(parent)

	if bind, ok := checkMatchStylePlaceholder(s); ok {
		if _, exists := parentBindSet[bind]; exists {
			return nil, errors.Errorf("duplicated bind parameter %q in position %d", bind, s.Pos.Offset)
		}
		return &placeholderTree{
			baseTree: baseTree{
				parent:  parent,
				segment: s,
			},
			bind: bind,
		}, nil
	}

	if bind, capture, ok := checkMatchStyleAll(s); ok {
		if _, exists := parentBindSet[bind]; exists {
			return nil, errors.Errorf("duplicated bind parameter %q in position %d", bind, s.Pos.Offset)
		}

		// One route can only have at most one match all style for subtree.
		ancestor := parent
		for ancestor != nil {
			if ancestor.getMatchStyle() == matchStyleAll {
				return nil, errors.Errorf("duplicated match all style in position %d", s.Pos.Offset)
			}
			ancestor = ancestor.getParent()
		}

		return &matchAllTree{
			baseTree: baseTree{
				parent:  parent,
				segment: s,
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

	return &regexTree{
		baseTree: baseTree{
			parent:  parent,
			segment: s,
		},
		regexp: re,
		binds:  binds,
	}, nil
}

// NewTree creates and returns a root tree.
func NewTree() Tree {
	return &baseTree{}
}

// AddRoute adds a new route to the tree and associates the given handler.
func AddRoute(t Tree, r *Route, h Handler) (Leaf, error) {
	if r == nil || len(r.Segments) == 0 {
		return nil, errors.New("cannot add empty route")
	}
	return addNextSegment(t, r, 0, h)
}

// matchLeaf returns the matched leaf and true if any leaf of the tree matches
// the given segment.
func (t *baseTree) matchLeaf(segment string, params Params, header http.Header) (Leaf, bool) {
	for _, l := range t.leaves {
		ok := l.match(segment, params, header)
		if ok {
			return l, true
		}
	}
	return nil, false
}

// matchSubtree returns the matched leaf and true if any subtree or leaf of the
// tree matches the given segment.
func (t *baseTree) matchSubtree(path, segment string, next int, params Params, header http.Header) (Leaf, bool) {
	for _, st := range t.subtrees {
		if st.getMatchStyle() == matchStyleAll {
			leaf, ok := st.(*matchAllTree).matchAll(path, segment, next, params, header)
			if ok {
				return leaf, true
			}

			// Any list of subtrees only has at most one match all style as the last
			// element. Therefore both "break" and "continue" have the same effect, but
			// using "break" here to be explicit.
			break
		}

		ok := st.match(segment, params)
		if !ok {
			continue
		}

		leaf, ok := st.matchNextSegment(path, next, params, header)
		if !ok {
			continue
		}
		return leaf, true
	}

	// Fall back to match all leaf of the tree.
	if len(t.leaves) > 0 {
		leaf := t.leaves[len(t.leaves)-1]
		if leaf.getMatchStyle() != matchStyleAll {
			return nil, false
		}

		ok := leaf.(*matchAllLeaf).matchAll(path, segment, next, params, header)
		if ok {
			return leaf, ok
		}
		return nil, false
	}
	return nil, false
}

func (t *baseTree) matchNextSegment(path string, next int, params Params, header http.Header) (Leaf, bool) {
	i := strings.Index(path[next:], "/")
	if i == -1 {
		return t.matchLeaf(path[next:], params, header)
	}
	return t.matchSubtree(path, path[next:next+i], next+i+1, params, header)
}

func (t *baseTree) Match(path string, header http.Header) (Leaf, Params, bool) {
	path = strings.TrimLeft(path, "/")
	params := make(Params)
	leaf, ok := t.matchNextSegment(path, 0, params, header)
	if !ok {
		return nil, nil, false
	}

	for k, v := range params {
		unescaped, err := url.PathUnescape(v)
		if err == nil {
			params[k] = unescaped
		}
	}
	return leaf, params, true
}
