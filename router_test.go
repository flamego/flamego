// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flamego/flamego/internal/route"
)

func TestRouter_Route(t *testing.T) {
	ctx := newMockContext()
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		ctx.params = Params(params)
		return ctx
	}
	r := newRouter(contextCreator)

	t.Run("register invalid HTTP method", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "unknown HTTP method:")
		}()

		r.Route("404", "/", nil)
	})

	t.Run("request with invalid HTTP method", func(t *testing.T) {
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("UNEXPECTED", "/", nil)
		assert.Nil(t, err)

		ctx.run_ = func() {}
		r.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	tests := []struct {
		routePath string
		method    string
		add       func(routePath string, handlers ...Handler) *Route
	}{
		{
			routePath: "/get",
			method:    http.MethodGet,
			add:       r.Get,
		},
		{
			routePath: "/patch",
			method:    http.MethodPatch,
			add:       r.Patch,
		},
		{
			routePath: "/post",
			method:    http.MethodPost,
			add:       r.Post,
		},
		{
			routePath: "/put",
			method:    http.MethodPut,
			add:       r.Put,
		},
		{
			routePath: "/delete",
			method:    http.MethodDelete,
			add:       r.Delete,
		},
		{
			routePath: "/options",
			method:    http.MethodOptions,
			add:       r.Options,
		},
		{
			routePath: "/head",
			method:    http.MethodHead,
			add:       r.Head,
		},
		{
			routePath: "/connect",
			method:    http.MethodConnect,
			add:       r.Connect,
		},
		{
			routePath: "/trace",
			method:    http.MethodTrace,
			add:       r.Trace,
		},
		{
			routePath: "/any",
			method:    http.MethodHead,
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
	ctx := newMockContext()
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		ctx.params = Params(params)
		return ctx
	}
	r := newRouter(contextCreator)

	r.Routes("/routes", "GET,POST", func() {})

	for _, m := range []string{http.MethodGet, http.MethodPost} {
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
	ctx := newMockContext()
	contextCreator := func(_ http.ResponseWriter, _ *http.Request, params route.Params, _ []Handler, _ urlPather) Context {
		ctx.params = Params(params)
		return ctx
	}

	t.Run("no auto head", func(t *testing.T) {
		r := newRouter(contextCreator)
		r.Get("/", func() {})

		gotRoute := ""
		ctx.run_ = func() { gotRoute = ctx.params["route"] }

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodHead, "/", nil)
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
		req, err := http.NewRequest(http.MethodHead, "/", nil)
		assert.Nil(t, err)

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "/", gotRoute)
	})
}

func TestRouter_DuplicatedRoutes(t *testing.T) {
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		return newMockContext()
	}
	r := newRouter(contextCreator)

	r.Get("/", func() {})

	defer func() {
		assert.Contains(t, recover(), "duplicated route")
	}()
	r.Get("/", func() {})
}

func TestRoute_Name(t *testing.T) {
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		return newMockContext()
	}
	r := newRouter(contextCreator)

	r.Get("/", func() {}).Name("home")

	t.Run("duplicated route name", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "duplicated route name:")
		}()
		r.Get("/home", func() {}).Name("home")
	})

	t.Run("empty route name", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "empty route name")
		}()
		r.Get("/404", func() {}).Name("")
	})
}

func TestRouter_URLPath(t *testing.T) {
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		return newMockContext()
	}
	r := newRouter(contextCreator)

	t.Run("name not exists", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "route with given name does not exist:")
		}()
		r.URLPath("404")
	})

	r.Get("/{owner}/{repo}/settings/?keys").Name("repo.settings")
	tests := []struct {
		name  string
		pairs []string
		want  string
	}{
		{
			name:  "good",
			pairs: []string{"owner", "flamego", "repo", "flamego"},
			want:  "/flamego/flamego/settings",
		},
		{
			name:  "missing last value",
			pairs: []string{"owner", "flamego", "repo"},
			want:  "/flamego/{repo}/settings",
		},
		{
			name:  "withOptional",
			pairs: []string{"owner", "flamego", "repo", "flamego", "withOptional", "true"},
			want:  "/flamego/flamego/settings/keys",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := r.URLPath("repo.settings", test.pairs...)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestRouter_Group(t *testing.T) {
	ctx := newMockContext()
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		ctx.params = Params(params)
		return ctx
	}
	r := newRouter(contextCreator)

	r.Get("/home")
	r.Group("/api", func() {
		r.Group("/v1", func() {
			r.Get("/users", func() {})
		})
		r.Get("/v2/users", func() {})
	})
	r.Get("/repos")

	routes := []string{
		"/home",
		"/api/v1/users",
		"/api/v2/users",
		"/repos",
	}
	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			gotRoute := ""
			ctx.run_ = func() { gotRoute = ctx.params["route"] }

			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, route, nil)
			assert.Nil(t, err)

			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Equal(t, route, gotRoute)
		})
	}
}

func TestComboRoute(t *testing.T) {
	ctx := newMockContext()
	contextCreator := func(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
		ctx.params = Params(params)
		return ctx
	}
	r := newRouter(contextCreator)

	r.Combo("/").
		Get(func() {}).
		Patch(func() {}).
		Post(func() {}).
		Put(func() {}).
		Delete(func() {}).
		Options(func() {}).
		Head(func() {}).
		Connect(func() {}).
		Trace(func() {})

	for _, m := range httpMethods {
		t.Run(m, func(t *testing.T) {
			gotRoute := ""
			ctx.run_ = func() { gotRoute = ctx.params["route"] }

			resp := httptest.NewRecorder()
			req, err := http.NewRequest(m, "/", nil)
			assert.Nil(t, err)

			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Equal(t, "/", gotRoute)
		})
	}
}
