package httpbulb

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func Base64DecodeHandle(w http.ResponseWriter, r *http.Request) {
	// Decode the base64 encoded string
	value := chi.URLParam(r, "value")

	decoded, err := base64.URLEncoding.DecodeString(value)

	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return

	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(decoded)

}

// StreamNMessagesHandle streams N json messages
func StreamNMessagesHandle(w http.ResponseWriter, r *http.Request) {
	// Stream N messages
	nParam := chi.URLParam(r, "n")
	n, err := strconv.Atoi(nParam)
	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	n = min(n, 100)

	resp := &StreamResponse{
		Args:    r.URL.Query(),
		Headers: r.Header,
		Origin:  getIP(r),
		URL:     getAbsoluteURL(r),
	}

	flusher := w.(http.Flusher)

	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	enc := json.NewEncoder(w)

	for i := 0; i < n; i++ {
		resp.ID = i
		enc.Encode(resp)
		flusher.Flush()
	}

}

// DelayHandle returns the same response as the MethodsHandle, but with a delay
func DelayHandle(w http.ResponseWriter, r *http.Request) {
	delayParam := chi.URLParam(r, "delay")
	d, err := strconv.Atoi(delayParam)
	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	d = min(d, 10)

	delay := time.Duration(d) * time.Second

	// Delay for d milliseconds
	<-time.After(delay)

	resp, err := newMethodResponse(r)

	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, http.StatusOK, resp)

}
