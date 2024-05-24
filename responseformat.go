package httpbulb

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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
