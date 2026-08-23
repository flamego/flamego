// Copyright 2026 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package method provides type-safe HTTP method sets for route registration.
package method

// Set is a set of HTTP methods.
type Set uint16

// HTTP method flags. Combine multiple methods with the bitwise OR operator.
const (
	Get Set = 1 << iota
	Post
	Put
	Delete
	Patch
	Options
	Head
	Connect
	Trace

	// All contains every HTTP method supported by Flamego.
	All = Get | Post | Put | Delete | Patch | Options | Head | Connect | Trace
)
