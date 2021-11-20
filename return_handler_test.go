// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReturnHandler(t *testing.T) {
	tests := []struct {
		name     string
		handler  Handler
		wantCode int
		wantBody string
	}{
		{
			name: "(int, string)",
			handler: func() (int, string) {
				return http.StatusTeapot, "i'm a teapot"
			},
			wantCode: http.StatusTeapot,
			wantBody: "i'm a teapot",
		},
		{
			name: "(int, []byte)",
			handler: func() (int, []byte) {
				return http.StatusTeapot, []byte("i'm a teapot")
			},
			wantCode: http.StatusTeapot,
			wantBody: "i'm a teapot",
		},
		{
			name: "(int, error)",
			handler: func() (int, error) {
				return http.StatusForbidden, errors.New("teapot on the phone")
			},
			wantCode: http.StatusForbidden,
			wantBody: "teapot on the phone",
		},
		{
			name: "string",
			handler: func() string {
				return "my boss, my hero"
			},
			wantCode: http.StatusOK,
			wantBody: "my boss, my hero",
		},
		{
			name: "[]byte",
			handler: func() []byte {
				return []byte("my boss, my hero")
			},
			wantCode: http.StatusOK,
			wantBody: "my boss, my hero",
		},
		{
			name: "error",
			handler: func() error {
				return errors.New("teapot on the phone")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "teapot on the phone",
		},
		{
			name: "nil-error",
			handler: func() error {
				return nil
			},
			wantCode: http.StatusOK,
			wantBody: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := New()
			f.Get("/", test.handler)

			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)

			assert.Equal(t, test.wantCode, resp.Code)
			assert.Equal(t, test.wantBody, resp.Body.String())
		})
	}
}
