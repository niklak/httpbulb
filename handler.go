package httpbulb

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Delete("/delete", methodsHandle)
	r.Get("/get", methodsHandle)
	r.Patch("/patch", methodsHandle)
	r.Post("/post", methodsHandle)
	r.Put("/put", methodsHandle)

	// TODO: add a handler that accepts a sequence of status codes
	// and returns a random status code from it.
	r.Handle("/status/{code:[1-5][0-9][0-9]}", http.HandlerFunc(statusCodeHandle))

	return r
}
