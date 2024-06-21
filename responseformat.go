package httpbulb

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/andybalholm/brotli"
)

const angryASCII = `
(╯°□°）╯︵ ┻━┻
YOU SHOULDN'T BE HERE
`

// GzipHandle is the handler that returns a response compressed with gzip
func GzipHandle(w http.ResponseWriter, r *http.Request) {

	var err error

	response, err := newMethodResponse(r)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Gzipped = true

	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)

	enc := json.NewEncoder(gz)
	if err = enc.Encode(response); err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gz.Close()

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, buf)
}

// DeflateHandle is the handler that returns a response compressed with zlib
func DeflateHandle(w http.ResponseWriter, r *http.Request) {

	var err error

	response, err := newMethodResponse(r)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Deflated = true

	buf := new(bytes.Buffer)
	zl := zlib.NewWriter(buf)

	enc := json.NewEncoder(zl)
	if err = enc.Encode(response); err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	zl.Close()

	w.Header().Set("Content-Encoding", "deflate")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, buf)
}

// BrotliHandle is the handler that returns a response compressed with brotli
func BrotliHandle(w http.ResponseWriter, r *http.Request) {

	var err error

	response, err := newMethodResponse(r)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Brotli = true

	buf := new(bytes.Buffer)
	br := brotli.NewWriter(buf)

	enc := json.NewEncoder(br)
	if err = enc.Encode(response); err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	br.Close()

	w.Header().Set("Content-Encoding", "br")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, buf)
}

// RobotsHandle returns the `text/plain` content for the robots.txt
func RobotsHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	body := []byte(`
User-agent: *
Disallow: /deny
	`)
	w.Write(body)
}

// DenyHandle  returns a page denied by robots.txt rules.
func DenyHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(angryASCII))
}

// Utf8SampleHandle serves the utf8-encoded file
func Utf8SampleHandle(w http.ResponseWriter, r *http.Request) {
	serveFileFS(w, r, assetsFS, "assets/utf8.html")
}

// HtmlSampleHandle serves the html file
func HtmlSampleHandle(w http.ResponseWriter, r *http.Request) {
	serveFileFS(w, r, assetsFS, "assets/moby.html")
}

// JSONSampleHandle serves the json file
func JSONSampleHandle(w http.ResponseWriter, r *http.Request) {
	serveFileFS(w, r, assetsFS, "assets/sample.json")
}

// XMLSampleHandle serves the xml file
func XMLSampleHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	serveFileFS(w, r, assetsFS, "assets/sample.xml")
}
