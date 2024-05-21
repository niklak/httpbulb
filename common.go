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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err})
}
