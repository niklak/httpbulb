package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	schemeHttp  = "http"
	schemeHttps = "https"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Delete("/delete", methodsHandle)
	r.Get("/get", methodsHandle)
	r.Patch("/patch", methodsHandle)
	r.Post("/post", methodsHandle)
	r.Put("/put", methodsHandle)

	return r
}

func methodsHandle(w http.ResponseWriter, r *http.Request) {

	// TODO: add json support

	defer r.Body.Close()

	var err error

	var body []byte

	response := MethodsResponse{
		Args:    r.URL.Query(),
		Headers: r.Header,
		Origin:  r.RemoteAddr,
		URL:     getAbsoluteURL(r),
	}
	ct := r.Header.Get("Content-Type")
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if strings.HasPrefix(ct, "multipart/form-data") {
			if err = r.ParseMultipartForm(64 << 20); err != nil {
				JsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if r.MultipartForm != nil && r.MultipartForm.File != nil {
				files := make(map[string][]string)
				for k, f := range r.MultipartForm.File {
					for _, fileHeader := range f {
						file, err := fileHeader.Open()
						if err != nil {
							JsonError(w, err.Error(), http.StatusInternalServerError)
							return
						}
						fbody, err := io.ReadAll(file)
						if err != nil {
							JsonError(w, err.Error(), http.StatusInternalServerError)
							return
						}
						files[k] = append(files[k], string(fbody))
					}
				}
				response.Files = files
				response.Form = r.Form

			}
		} else if ct == "application/x-www-form-urlencoded" {
			if err = r.ParseForm(); err != nil {
				JsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response.Form = r.Form
		} else {
			body, err = io.ReadAll(r.Body)
			if err != nil {
				JsonError(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			response.Data = string(body)
		}
	}

	response.JSON = nil

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getAbsoluteURL(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}

	var scheme string
	if r.TLS == nil {
		scheme = schemeHttp
	} else {
		scheme = schemeHttps
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.RequestURI())
}

func JsonError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err})
}
