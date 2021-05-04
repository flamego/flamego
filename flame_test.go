// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Make sure Run doesn't blow up
func TestFlame_Run(t *testing.T) {
	var fs []*Flame

	f1 := NewWithLogger(io.Discard)
	go f1.Run()
	fs = append(fs, f1)

	f2 := NewWithLogger(io.Discard)
	go f2.Run(4002)
	fs = append(fs, f2)

	f3 := NewWithLogger(io.Discard)
	go f3.Run("0.0.0.0", 4003)
	fs = append(fs, f3)

	_ = os.Setenv("FLAMEGO_ADDR", ":4001")
	f4 := NewWithLogger(io.Discard)
	go f4.Run("0.0.0.0")
	fs = append(fs, f4)

	time.Sleep(1 * time.Second)

	for _, f := range fs {
		f.Stop()
	}
}

func TestFlame_Before(t *testing.T) {
	f := New()

	var buf bytes.Buffer
	f.Before(func(http.ResponseWriter, *http.Request) bool {
		buf.WriteString("foo")
		return false
	})
	f.Before(func(http.ResponseWriter, *http.Request) bool {
		buf.WriteString("bar")
		return true
	})
	f.Get("/", func() {
		buf.WriteString("boom")
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "foobar", buf.String())
}
