package httpbulb

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

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
