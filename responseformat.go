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
