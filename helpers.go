package httpbulb

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

const (
	schemeHttp  = "http"
	schemeHttps = "https"
)

var trueClientIP = "True-Client-IP"
var xForwardedFor = "X-Forwarded-For"
var xRealIP = "X-Real-IP"

//go:embed assets/*
var assetsFS embed.FS

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

	if realIP := r.Header.Get(xRealIP); realIP != "" {
		ip = realIP
	} else if clientIP := r.Header.Get(trueClientIP); clientIP != "" {
		ip = clientIP
	} else if forwardedFor := r.Header.Get(xForwardedFor); forwardedFor != "" {
		ip = strings.TrimSpace(strings.SplitN(forwardedFor, ",", 2)[0])
	} else {
		ip = r.RemoteAddr
	}
	return ip
}
func getRequestHeader(r *http.Request) http.Header {
	h := r.Header.Clone()
	h.Set("Host", r.Host)
	return h
}

func writeJsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// serveFileFS serves a file from the given filesystem
// this is a replacement of http.ServeFileFS because go 1.18 doesn't this function.
func serveFileFS(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	fs := http.FS(fsys)
	f, err := fs.Open(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, name, fi.ModTime(), f)
}

func setCookie(w http.ResponseWriter, name, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  value,
		Secure: secure,
		Path:   "/",
	})
}

func getCookie(r *http.Request, name string) (cookieValue string) {

	if cookie, err := r.Cookie(name); err == nil {
		cookieValue = cookie.Value
	}
	return
}
