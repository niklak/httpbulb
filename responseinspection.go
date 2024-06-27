package httpbulb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%s", value))

	MethodsHandle(w, r)

}

// EtagHandle assumes the resource has the given etag and responds to If-None-Match and If-Match headers appropriately.
func EtagHandle(w http.ResponseWriter, r *http.Request) {

	etag := chi.URLParam(r, "etag")

	ifNoneMatch := r.Header.Get("If-None-Match")
	ifMatch := r.Header.Get("If-Match")

	if ifNoneMatch != "" {
		if strings.Contains(ifNoneMatch, etag) || strings.Contains(ifNoneMatch, "*") {
			w.Header().Set("ETag", etag)
			w.WriteHeader(http.StatusNotModified)
			return

		}
	} else if ifMatch != "" {
		if !strings.Contains(ifMatch, etag) && !strings.Contains(ifMatch, "*") {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
	}
	w.Header().Set("ETag", etag)
	MethodsHandle(w, r)
}

// ResponseHeadersHandle returns the response headers as JSON response
func ResponseHeadersHandle(w http.ResponseWriter, r *http.Request) {

	argHeaders := r.URL.Query()

	for k, vv := range argHeaders {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseHeaders := w.Header()
	json.NewEncoder(w).Encode(responseHeaders)
}
