package api

import (
	"net/http"

	"github.com/alperkirkus/fintech-backend/internal/middleware"
)

type Router struct {
	mux         *http.ServeMux
	middlewares []middleware.Middleware
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) Use(m middleware.Middleware) {
	r.middlewares = append(r.middlewares, m)
}

func (r *Router) Handle(pattern string, handler http.Handler, extra ...middleware.Middleware) {
	if len(extra) > 0 {
		handler = middleware.Chain(extra...)(handler)
	}
	r.mux.Handle(pattern, handler)
}

func (r *Router) HandleFunc(pattern string, fn http.HandlerFunc, extra ...middleware.Middleware) {
	r.Handle(pattern, fn, extra...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	middleware.Chain(r.middlewares...)(r.mux).ServeHTTP(w, req)
}
