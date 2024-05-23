package httpbulb

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Delete("/delete", MethodsHandle)
	r.Get("/get", MethodsHandle)
	r.Patch("/patch", MethodsHandle)
	r.Post("/post", MethodsHandle)
	r.Put("/put", MethodsHandle)

	r.Get("/headers", http.HandlerFunc(HeadersHandle))
	r.Get("/ip", http.HandlerFunc(IpHandle))
	r.Get("/user-agent", http.HandlerFunc(UserAgentHandle))

	// TODO: add a handler that accepts a sequence of status codes
	// and returns a random status code from it.
	r.Handle("/status/{code:[1-5][0-9][0-9]}", http.HandlerFunc(statusCodeHandle))
	r.Handle("/anything", http.HandlerFunc(MethodsHandle))
	r.Handle("/anything/{anything}", http.HandlerFunc(MethodsHandle))

	//TODO: add documentation

	return r
}
