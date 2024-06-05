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

	r.Get("/robots.txt", http.HandlerFunc(RobotsHandle))

	r.Get("/gzip", http.HandlerFunc(GzipHandle))
	r.Get("/deflate", http.HandlerFunc(DeflateHandle))
	r.Get("/brotli", http.HandlerFunc(BrotliHandle))

	// TODO: add a handler that accepts a sequence of status codes
	// and returns a random status code from it.
	r.Handle("/status/{code:[1-5][0-9][0-9]}", http.HandlerFunc(statusCodeHandle))
	r.Handle("/anything", http.HandlerFunc(MethodsHandle))
	r.Handle("/anything/{anything}", http.HandlerFunc(MethodsHandle))

	r.Get("/basic-auth/{user}/{passwd}", http.HandlerFunc(BasicAuthHandle))
	r.Get("/bearer", http.HandlerFunc(BearerAuthHandle))

	r.Get("/base64/{value}", http.HandlerFunc(Base64DecodeHandle))
	r.Get("/stream/{n:[0-9]+}", http.HandlerFunc(StreamNMessagesHandle))
	r.Get("/stream-bytes/{n:[0-9]+}", http.HandlerFunc(StreamRandomBytesHandle))
	r.Get("/bytes/{n:[0-9]+}", http.HandlerFunc(RandomBytesHandle))
	r.Get("/drip", http.HandlerFunc(DripHandle))
	r.Get("/uuid", http.HandlerFunc(UUIDHandle))
	r.Handle("/delay/{delay:[0-9]+}", http.HandlerFunc(DelayHandle))

	r.Get("/cookies", http.HandlerFunc(CookiesHandle))
	r.Get("/cookies-list", http.HandlerFunc(CookiesListHandle))
	r.Get("/cookies/set", http.HandlerFunc(SetCookiesHandle))
	r.Get("/cookies/set/{name}/{value}", http.HandlerFunc(SetCookieHandle))
	r.Get("/cookies/delete", http.HandlerFunc(DeleteCookiesHandle))

	//TODO: add documentation

	return r
}
