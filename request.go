// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"io"
	"net/http"
)

// Request is a wrapper of http.Request with handy methods.
type Request struct {
	*http.Request
}

// Body returns a RequestBody for the request.
func (r *Request) Body() *RequestBody {
	return &RequestBody{reader: r.Request.Body}
}

// RequestBody is a wrapper of http.Request.Body with handy methods.
type RequestBody struct {
	reader io.ReadCloser
}

// Bytes reads and returns the content of request body in bytes.
func (r *RequestBody) Bytes() ([]byte, error) {
	return io.ReadAll(r.reader)
}

// String reads and returns content of request body in string.
func (r *RequestBody) String() (string, error) {
	data, err := r.Bytes()
	return string(data), err
}

// ReadCloser returns a ReadCloser of request body.
func (r *RequestBody) ReadCloser() io.ReadCloser {
	return r.reader
}
