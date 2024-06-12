package httpbulb

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	schemeHttp  = "http"
	schemeHttps = "https"
)

//go:embed assets/*
var assetsFS embed.FS

func getURLScheme(r *http.Request) string {
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if r.TLS != nil {
		return schemeHttps
	}
	return schemeHttp

}

func getAbsoluteURL(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}
	scheme := getURLScheme(r)
	absURL := *r.URL
	absURL.Scheme = scheme
	absURL.Host = r.Host
	return absURL.String()
}

func getIP(r *http.Request) string {
	var ip string
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = strings.TrimSpace(strings.SplitN(forwardedFor, ",", 2)[0])
	} else {
		ip = r.RemoteAddr
	}
	return ip
}

func TextError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintln(w, err)
}

func JsonError(w http.ResponseWriter, err string, code int) {

	if err == "" {
		err = http.StatusText(code)
	}
	writeJsonResponse(w, code, map[string]string{"error": err})
}

func writeJsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
