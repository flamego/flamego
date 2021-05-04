// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		r := Request{
			Request: &http.Request{
				Body: io.NopCloser(bytes.NewReader([]byte("foobar"))),
			},
		}

		body, err := r.Body().Bytes()
		assert.Nil(t, err)
		assert.Equal(t, []byte("foobar"), body)
	})

	t.Run("string", func(t *testing.T) {
		r := Request{
			Request: &http.Request{
				Body: io.NopCloser(bytes.NewReader([]byte("foobar"))),
			},
		}

		body, err := r.Body().String()
		assert.Nil(t, err)
		assert.Equal(t, "foobar", body)
	})

	t.Run("ReadCloser", func(t *testing.T) {
		r := Request{
			Request: &http.Request{
				Body: io.NopCloser(bytes.NewReader([]byte("foobar"))),
			},
		}

		body, err := io.ReadAll(r.Body().ReadCloser())
		assert.Nil(t, err)
		assert.Equal(t, []byte("foobar"), body)
	})
}
