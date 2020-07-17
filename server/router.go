package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mleone896/inventory/db"
)

// APIContext ...
type APIContext struct {
	dao *db.DataObj
}

// LoadHandlers returns a new router with the available endpoints
func (ctx APIContext) LoadHandlers() *mux.Router {

	r := mux.NewRouter()
	v1 := r.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/new_host", WithLogging(ctx.NewTagsReq, "NewTagsRequest")).Methods("POST")
	v1.HandleFunc("/host/{id}", WithLogging(ctx.ListHostAttrsByColor, "ListHostAttrsByColor")).Methods("GET")
	v1.HandleFunc("/colors", WithLogging(ctx.ListColors, "ListColors")).Methods("GET")

	return r
}

// New ...
func New(opts ...func(*APIContext)) *APIContext {

	actx := &APIContext{}

	for _, opt := range opts {
		opt(actx)
	}

	return actx
}

// WithLogging decorates the request with some basic logs
func WithLogging(f http.HandlerFunc, handler string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("starting method %s call to %s handler", r.Method, handler)
		f(w, r)
		log.Printf("finishing method call to handler %s", r.Method)
	})
}

// WithDAO sets the data access object
func WithDAO(dao *db.DataObj) func(*APIContext) {
	return func(actx *APIContext) {
		actx.dao = dao
	}

}
