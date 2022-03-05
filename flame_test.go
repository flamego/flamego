// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClassic(_ *testing.T) {
	_ = Classic() // For the sake of code coverage
}

// Make sure Run doesn't blow up
func TestFlame_Run(_ *testing.T) {
	_ = os.Setenv("FLAMEGO_ADDR", "0.0.0.0:4001")
	f := NewWithLogger(&bytes.Buffer{})
	go f.Run(4002)

	time.Sleep(1 * time.Second)

	f.Stop()
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
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "foobar", buf.String())
}

func TestFlame_ServeHTTP(t *testing.T) {
	var buf bytes.Buffer
	f := New()
	f.Use(func(c Context) {
		buf.WriteString("foo")
		c.Next()
		buf.WriteString("ban")
	})
	f.Use(func(c Context) {
		buf.WriteString("bar")
		c.Next()
		buf.WriteString("baz")
	})
	f.Get("/", func() {})
	f.Action(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteString("bat")
		w.WriteHeader(http.StatusBadRequest)
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "foobarbatbazban", buf.String())
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestFlame_Handlers(t *testing.T) {
	var buf bytes.Buffer
	batman := func() {
		buf.WriteString("batman!")
	}

	f := New()
	f.Use(func(c Context) {
		buf.WriteString("foo")
		c.Next()
		buf.WriteString("ban")
	})
	f.Handlers(
		batman,
		batman,
		batman,
	)

	f.Get("/", func() {})
	f.Action(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteString("bat")
		w.WriteHeader(http.StatusBadRequest)
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "batman!batman!batman!bat", buf.String())
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestFlame_EarlyWrite(t *testing.T) {
	var buf bytes.Buffer
	f := New()
	f.Use(func(w http.ResponseWriter) {
		buf.WriteString("foobar")
		_, _ = w.Write([]byte("Hello world"))
	})
	f.Use(func() {
		buf.WriteString("bat")
	})
	f.Get("/", func() {})
	f.Action(func(w http.ResponseWriter) {
		buf.WriteString("baz")
		w.WriteHeader(http.StatusBadRequest)
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "foobar", buf.String())
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestFlame_NoRace(t *testing.T) {
	f := New()
	handlers := []Handler{func() {}, func() {}}
	// Ensure append will not reallocate alice that triggers the race condition
	f.handlers = handlers[:1]
	f.Get("/", func() {})
	for i := 0; i < 2; i++ {
		go func() {
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			resp := httptest.NewRecorder()
			assert.Nil(t, err)

			f.ServeHTTP(resp, req)
		}()
	}
}

func TestEnv(t *testing.T) {
	defer SetEnv(EnvTypeDev)
	envs := []EnvType{
		EnvTypeDev,
		EnvTypeProd,
		EnvTypeTest,
	}
	for _, env := range envs {
		SetEnv(env)
		assert.Equal(t, env, Env())
	}
}
