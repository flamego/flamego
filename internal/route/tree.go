// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

// Tree is a tree derived from a segment.
type Tree interface {
	// AddRoute adds a new route to the tree and associates given handler to it.
	AddRoute(r *Route, h Handler) (Leaf, error)

	// getParent returns the parent tree. The root tree does not have parent.
	getParent() Tree
	// getSegment returns the segment that the tree is derived from.
	getSegment() *Segment
	// getMatchStyle returns the match style of the tree.
	getMatchStyle() MatchStyle
	// addNextSegment adds next segment of the route to the tree.
	addNextSegment(t Tree, r *Route, next int, h Handler) (Leaf, error)
	// addLeaf adds a new leaf from the given segment.
	addLeaf(t Tree, r *Route, s *Segment, h Handler) (Leaf, error)
	// addSubtree adds a new subtree from next segment of the route.
	addSubtree(t Tree, r *Route, next int, h Handler) (Leaf, error)
	// getSubtrees returns the list of direct subtrees.
	getSubtrees() []Tree
	// getLeaves returns the list of direct leaves.
	getLeaves() []Leaf
	// setSubtrees sets the list of direct subtrees.
	setSubtrees(subtrees []Tree)
	// setLeaves sets the list of direct leaves.
	setLeaves(leaves []Leaf)
}

// baseTree contains common fields for any tree.
type baseTree struct {
	parent   Tree     // The parent tree.
	segment  *Segment // The segment that the tree is derived from.
	subtrees []Tree   // The list of direct subtrees.
	leaves   []Leaf   // The list of direct leaves.
}

func (t *baseTree) getParent() Tree {
	return t.parent
}

func (t *baseTree) getSegment() *Segment {
	return t.segment
}

func (t *baseTree) getMatchStyle() MatchStyle {
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

// Params is a set of bind parameters extracted from the URL.
type Params map[string]string

// Handler is a function that can be registered to a route for handling HTTP
// requests.
type Handler func(http.ResponseWriter, *http.Request, Params)

func (*baseTree) addLeaf(t Tree, r *Route, s *Segment, h Handler) (Leaf, error) {
	leaves := t.getLeaves()
	for _, l := range leaves {
		if l.getSegment().String() == s.String() {
			return l, nil
		}
	}

	leaf, err := newLeaf(t, r, s, h)
	if err != nil {
		return nil, errors.Wrap(err, "new leaf")
	}

	if leaf.getSegment().Optional {
		parent := leaf.getParent()
		if parent.getParent() != nil {
			_, err = parent.getParent().addLeaf(t, r, parent.getSegment(), h)
			if err != nil {
				return nil, errors.Wrap(err, "add optional leaf to grandparent")
			}
		} else {
			_, err = parent.addLeaf(t, r, parent.getSegment(), h)
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

func (*baseTree) addSubtree(t Tree, r *Route, next int, h Handler) (Leaf, error) {
	for _, st := range t.getSubtrees() {
		if st.getSegment().String() == r.Segments[next].String() {
			return st.addNextSegment(st, r, next+1, h)
		}
	}

	subtree, err := newTree(t, r.Segments[next])
	if err != nil {
		return nil, errors.Wrap(err, "new tree")
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

	return subtree.addNextSegment(subtree, r, next+1, h)
}

func (*baseTree) addNextSegment(t Tree, r *Route, next int, h Handler) (Leaf, error) {
	if len(r.Segments) <= next+1 {
		return t.addLeaf(t, r, r.Segments[next], h)
	}

	if r.Segments[next].Optional {
		return nil, errors.New("only the last segment can be optional")
	}
	return t.addSubtree(t, r, next, h)
}

func (t *baseTree) AddRoute(r *Route, h Handler) (Leaf, error) {
	if r == nil || len(r.Segments) == 0 {
		return nil, errors.New("cannot add empty route")
	}
	return t.addNextSegment(&staticTree{baseTree: *t}, r, 0, h)
}

// staticTree is a tree with a static match style.
type staticTree struct {
	baseTree
}

func (t *staticTree) getMatchStyle() MatchStyle {
	return matchStyleStatic
}

// regexTree is a tree with a regex match style.
type regexTree struct {
	baseTree
	regexp *regexp.Regexp // The regexp for the tree.
	binds  []string       // The list of bind parameters.
}

func (t *regexTree) getMatchStyle() MatchStyle {
	return matchStyleRegex
}

// placeholderTree is a tree with a placeholder match style.
type placeholderTree struct {
	baseTree
}

func (l *placeholderTree) getMatchStyle() MatchStyle {
	return matchStylePlaceholder
}

// placeholderTree is a tree with a match all style.
type matchAllTree struct {
	baseTree
	bind string // The name of the bind parameter.
}

func (l *matchAllTree) getMatchStyle() MatchStyle {
	return matchStyleAll
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

	if isMatchStylePlaceholder(s) {
		return &placeholderTree{
			baseTree: baseTree{
				parent:  parent,
				segment: s,
			},
		}, nil
	}

	if bind, ok := checkMatchStyleAll(s); ok {
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
			bind: bind,
		}, nil
	}

	// The only remaining style is regex
	re, binds, err := constructMatchStyleRegex(s)
	if err != nil {
		return nil, err
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
