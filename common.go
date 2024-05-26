package httpbulb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	schemeHttp  = "http"
	schemeHttps = "https"
)

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
	writeJsonResponse(w, code, map[string]string{"error": err})
}

func writeJsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
