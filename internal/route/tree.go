// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/pkg/errors"
)

// todo
type Tree struct {
	parent  *Tree
	segment *Segment
	leaves  []Leaf
}

// todo
func NewTree() *Tree {
	return &Tree{}
}

// todo
type Params map[string]string

// todo
type Handler func(http.ResponseWriter, *http.Request, Params)

// todo
func (t *Tree) addLeaf(r *Route, s *Segment, h Handler) (Leaf, error) {
	// todo: check if the same segment has been added already

	leaf, err := newLeaf(t, r, s, h)
	if err != nil {
		return nil, errors.Wrap(err, "new leaf")
	}

	if leaf.Optional() {
		parent := leaf.Parent()
		if parent.parent != nil {
			_, err = parent.parent.addLeaf(r, parent.segment, h)
			if err != nil {
				return nil, errors.Wrap(err, "add optional leaf to grandparent")
			}
		} else {
			_, err = parent.addLeaf(r, parent.segment, h)
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
func (t *Tree) addSubtree(r *Route, next int, h Handler) (Leaf, error) {
	return nil, nil // todo
}

// todo
func (t *Tree) addNextSegment(r *Route, next int, h Handler) (Leaf, error) {
	if len(r.Segments) >= next+1 {
		return t.addLeaf(r, r.Segments[next], h)
	}

	if r.Segments[next].Optional {
		return nil, errors.New("only the last segment can be optional")
	}
	return t.addSubtree(r, next+1, h)
}

// todo
func (t *Tree) AddRoute(r *Route, h Handler) (Leaf, error) {
	if r == nil || len(r.Segments) == 0 {
		return nil, errors.New("cannot add empty route")
	}
	return t.addNextSegment(r, 0, h)
}
