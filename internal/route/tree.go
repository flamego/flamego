// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

// todo
type Tree interface {
	// Parent returns the parent tree. The root tree does not have parent.
	Parent() Tree
	// Segment returns the segment that the tree is derived from.
	Segment() *Segment
	// MatchStyle returns the match style of the tree.
	MatchStyle() MatchStyle
	// todo
	addNextSegment(r *Route, next int, h Handler) (Leaf, error)
	// todo
	addLeaf(r *Route, s *Segment, h Handler) (Leaf, error)
}

// todo
type baseTree struct {
	parent   Tree     // The parent tree.
	segment  *Segment // The segment that the tree is derived from.
	subtrees []Tree   // The list of direct subtrees.
	leaves   []Leaf   // The list of direct leaves.
}

func (t *baseTree) Parent() Tree {
	return t.parent
}

func (t *baseTree) Segment() *Segment {
	return t.segment
}

func (t *baseTree) MatchStyle() MatchStyle {
	return matchStyleNone
}

// todo
type Params map[string]string

// todo
type Handler func(http.ResponseWriter, *http.Request, Params)

func (t *baseTree) addLeaf(r *Route, s *Segment, h Handler) (Leaf, error) {
	for _, l := range t.leaves {
		if l.Segment().String() == s.String() {
			return l, nil
		}
	}

	leaf, err := newLeaf(t, r, s, h)
	if err != nil {
		return nil, errors.Wrap(err, "new leaf")
	}

	if leaf.Segment().Optional {
		parent := leaf.Parent()
		if parent.Parent() != nil {
			_, err = parent.Parent().addLeaf(r, parent.Segment(), h)
			if err != nil {
				return nil, errors.Wrap(err, "add optional leaf to grandparent")
			}
		} else {
			_, err = parent.addLeaf(r, parent.Segment(), h)
			if err != nil {
				return nil, errors.Wrap(err, "add optional leaf to parent")
			}
		}
	}

	// Determine leaf position by the priority of match styles.
	i := 0
	for ; i < len(t.leaves); i++ {
		if leaf.MatchStyle() < t.leaves[i].MatchStyle() {
			break
		}
	}

	if i == len(t.leaves) {
		t.leaves = append(t.leaves, leaf)
	} else {
		t.leaves = append(t.leaves[:i], append([]Leaf{leaf}, t.leaves[i:]...)...)
	}
	return leaf, nil
}

// todo
func (t *baseTree) addSubtree(r *Route, next int, h Handler) (Leaf, error) {
	for _, st := range t.subtrees {
		if st.Segment().String() == r.Segments[next].String() {
			return st.addNextSegment(r, next+1, h)
		}
	}

	subtree, err := newTree(t, r.Segments[next])
	if err != nil {
		return nil, errors.Wrap(err, "new tree")
	}

	// Determine subtree position by the priority of match styles.
	i := 0
	for ; i < len(t.subtrees); i++ {
		if subtree.MatchStyle() < t.subtrees[i].MatchStyle() {
			break
		}
	}

	if i == len(t.subtrees) {
		t.subtrees = append(t.subtrees, subtree)
	} else {
		t.subtrees = append(t.subtrees[:i], append([]Tree{subtree}, t.subtrees[i:]...)...)
	}
	return subtree.addNextSegment(r, next+1, h)
}

func (t *baseTree) addNextSegment(r *Route, next int, h Handler) (Leaf, error) {
	if len(r.Segments) >= next+1 {
		return t.addLeaf(r, r.Segments[next], h)
	}

	if r.Segments[next].Optional {
		return nil, errors.New("only the last segment can be optional")
	}
	return t.addSubtree(r, next+1, h)
}

// todo
func (t *baseTree) AddRoute(r *Route, h Handler) (Leaf, error) {
	if r == nil || len(r.Segments) == 0 {
		return nil, errors.New("cannot add empty route")
	}
	return t.addNextSegment(r, 0, h)
}

// staticTree is a tree with a static match style.
type staticTree struct {
	baseTree
}

func (t *staticTree) MatchStyle() MatchStyle {
	return matchStyleStatic
}

// regexTree is a tree with a regex match style.
type regexTree struct {
	baseTree
	regexp *regexp.Regexp // The regexp for the tree.
	binds  []string       // The list of bind parameters.
}

func (t *regexTree) MatchStyle() MatchStyle {
	return matchStyleRegex
}

// placeholderTree is a tree with a placeholder match style.
type placeholderTree struct {
	baseTree
}

func (l *placeholderTree) MatchStyle() MatchStyle {
	return matchStylePlaceholder
}

// placeholderTree is a tree with a match all style.
type matchAllTree struct {
	baseTree
	bind string // The name of the bind parameter.
}

func (l *matchAllTree) MatchStyle() MatchStyle {
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
			if ancestor.MatchStyle() == matchStyleAll {
				return nil, errors.Errorf("duplicated match all style in position %d", s.Pos.Offset)
			}
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
