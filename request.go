// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
)

// Request is a wrapper of http.Request with handy methods.
type Request struct {
	*http.Request
}
