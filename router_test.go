// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flamego/flamego/internal/route"
	"github.com/stretchr/testify/assert"
)

func TestRouter_Route(t *testing.T) {
	ctx := &mockContext{}
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler) Context {
		ctx.params = params
		return ctx
	}
	r := newRouter(contextCreator)

	t.Run("invalid HTTP method", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "unknown HTTP method:")
		}()

		r.Route("404", "/", nil)
	})

	tests := []struct {
		routePath string
		method    string
		add       func(routePath string, handlers ...Handler) *Route
	}{
		{
			routePath: "/get",
			method:    "GET",
			add:       r.Get,
		},
		{
			routePath: "/patch",
			method:    "PATCH",
			add:       r.Patch,
		},
		{
			routePath: "/post",
			method:    "POST",
			add:       r.Post,
		},
		{
			routePath: "/put",
			method:    "PUT",
			add:       r.Put,
		},
		{
			routePath: "/delete",
			method:    "DELETE",
			add:       r.Delete,
		},
		{
			routePath: "/options",
			method:    "OPTIONS",
			add:       r.Options,
		},
		{
			routePath: "/head",
			method:    "HEAD",
			add:       r.Head,
		},
		{
			routePath: "/any",
			method:    "HEAD",
			add:       r.Any,
		},
	}
	for _, test := range tests {
		t.Run(test.routePath, func(t *testing.T) {
			test.add(test.routePath, func() {})

			gotRoute := ""
			ctx.run_ = func() { gotRoute = ctx.params["route"] }

			resp := httptest.NewRecorder()
			req, err := http.NewRequest(test.method, test.routePath, nil)
			assert.Nil(t, err)

			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Equal(t, test.routePath, gotRoute)
		})
	}
}

func TestRouter_Routes(t *testing.T) {
	ctx := &mockContext{}
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler) Context {
		ctx.params = params
		return ctx
	}
	r := newRouter(contextCreator)

	r.Routes("/routes", "GET,POST", func() {})

	for _, m := range []string{"GET", "POST"} {
		gotRoute := ""
		ctx.run_ = func() { gotRoute = ctx.params["route"] }

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(m, "/routes", nil)
		assert.Nil(t, err)

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "/routes", gotRoute)
	}
}

func TestRouter_AutoHead(t *testing.T) {
	ctx := &mockContext{}
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler) Context {
		ctx.params = params
		return ctx
	}

	t.Run("no auto head", func(t *testing.T) {
		r := newRouter(contextCreator)
		r.Get("/", func() {})

		gotRoute := ""
		ctx.run_ = func() { gotRoute = ctx.params["route"] }

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/", nil)
		assert.Nil(t, err)

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", gotRoute)
	})

	t.Run("has auto head", func(t *testing.T) {
		r := newRouter(contextCreator)
		r.AutoHead(true)
		r.Get("/", func() {})

		gotRoute := ""
		ctx.run_ = func() { gotRoute = ctx.params["route"] }

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/", nil)
		assert.Nil(t, err)

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "/", gotRoute)
	})
}

func TestRouter_DuplicatedRoutes(t *testing.T) {
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler) Context {
		return &mockContext{}
	}
	r := newRouter(contextCreator)

	defer func() {
		assert.Contains(t, recover(), "duplicated route")
	}()

	r.Get("/", func() {})
	r.Get("/", func() {})
}

func TestRouter_Name(t *testing.T) {
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler) Context {
		return &mockContext{}
	}
	r := newRouter(contextCreator)

	r.Get("/", func() {}).Name("home")

	defer func() {
		assert.Contains(t, recover(), "duplicated route name:")
	}()
	r.Get("/home", func() {}).Name("home")
}
