// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"reflect"
)

// ReturnHandler is a service that is called when a route handler returns
// something. The ReturnHandler is responsible for writing to the ResponseWriter
// based on the values that are passed into this function.
type ReturnHandler func(Context, []reflect.Value)
