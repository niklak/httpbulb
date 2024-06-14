package httpbulb

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CacheHandle returns a 304 if an If-Modified-Since header or If-None-Match is present.
// Returns the same as a `MethodsHandle` (/get) otherwise.
func CacheHandle(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("If-Modified-Since") != "" || r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Last-Modified", time.Now().Format(time.RFC1123))
	w.Header().Set("ETag", uuid.New().String())

	MethodsHandle(w, r)
}

// CacheControlHandle sets a Cache-Control header for n seconds
func CacheControlHandle(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")

	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%s", value))

	MethodsHandle(w, r)

}
