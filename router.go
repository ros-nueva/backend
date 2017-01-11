package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

type RequestFilter func(*http.Request) bool
type Middleware func(http.HandlerFunc) http.HandlerFunc
type Filters []RequestFilter
type Stack []Middleware
type Routes map[string]Path
type Router Group

type Path interface {
	Build(*mux.Router, string)
}

type IGroup interface {
	AmGroup()
	Build(*mux.Router, string)
}

type Group struct {
	Middleware Stack
	Paths Routes
}

type IRoute interface {
	AmRoute()
	Build(*mux.Router, string)
}

type Route struct {
	Handler    http.HandlerFunc
	Allow      Filters
	Middleware Stack
}

func (r Routes) Serve() http.Handler{
	router := mux.NewRouter().StrictSlash(true)
	for pattern, path := range r {
		path.Build(router, pattern)
	}
	return router
}

func Allow(f RequestFilter) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if f(r) {
				h(w, r)
			} else {
				// TODO
				w.WriteHeader(http.StatusForbidden)
			}
		}
	}
}

// Combine creates a RequestFilter that is the conjunction
// of all the RequestFilters in f.
func (f Filters) Combine() RequestFilter {
	return func(r *http.Request) bool {
		for _, filter := range f {
			if !filter(r) {
				return false
			}
		}
		return true
	}
}

// Apply returns an http.Handlerfunc that has had all of the
// Middleware functions in s, if any, to f.
func (s Stack) Apply(f http.HandlerFunc) http.HandlerFunc {
	g := f
	for _, middleware := range s {
		g = middleware(g)
	}
	return g
}

// Builds the endpoint described by e, by applying
// access restrictions and other middleware.
func (r Route) Build(parent *mux.Router, prefix string) {
	allowFilter := r.Allow.Combine()
	restricted := Allow(allowFilter)(r.Handler)
	parent.HandleFunc(prefix, r.Middleware.Apply(restricted))
}

func (r Route) AmRoute() {}

func (g Group) Build(parent *mux.Router, prefix string) {
	router := parent.PathPrefix(prefix).Subrouter()
	for pattern, path := range g.Paths { // api/users/{id}/
		switch path.(type) {
		case IRoute:
			path.Build(router, pattern)
		case IGroup:
			path.Build(parent, prefix + pattern)
		}
	}
}

func (g Group) AmGroup() {}
