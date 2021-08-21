// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flamego/flamego/inject"
)

func TestLogger(t *testing.T) {
	t.Run("fast invoker", func(t *testing.T) {
		assert.True(t, inject.IsFastInvoker(Logger()))
	})

	f := NewWithLogger(&bytes.Buffer{})
	f.Use(Logger())
	f.Get("/{code}", func(c Context) (int, string) {
		code := c.ParamInt("code")
		return code, http.StatusText(code)
	})
	codes := []int{
		http.StatusOK, http.StatusCreated, http.StatusAccepted,
		http.StatusMovedPermanently, http.StatusFound,
		http.StatusNotModified,
		http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}
	for _, code := range codes {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d", code), nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)

			assert.Equal(t, code, resp.Code)
		})
	}
}
