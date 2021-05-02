// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute_String(t *testing.T) {
	parser, err := NewParser()
	assert.Nil(t, err)

	routes := []string{
		"/webapi",
		"/webapi/users",
		"/webapi/users/?{id}",
		"/{name}",
		"/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}",
		"/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}/{year: regex2}-{month-day}",
		"/webapi/{name-1}/{name-2: /[a-z0-9]{7, 40}/}/{year: regex2}-{month-day}/{**: **, capture: 3}",
	}
	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			r, err := parser.Parse(route)
			assert.Nil(t, err)
			assert.Equal(t, route, r.String())
		})
	}
}
