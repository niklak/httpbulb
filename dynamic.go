package httpbulb

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Base64DecodeHandle decodes a base64 encoded string and returns the decoded value
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
	totalMessages, err := strconv.Atoi(nParam)
	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	totalMessages = min(totalMessages, 100)

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

	for i := 0; i < totalMessages; i++ {
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

// RandomBytesHandle returns `n` random bytes generated with given `seed`.
func RandomBytesHandle(w http.ResponseWriter, r *http.Request) {
	// total bytes
	totalBytes, err := extractTotalBytes(r)
	if err != nil {
		TextError(w, "n: bad parameter", http.StatusBadRequest)
		return
	}

	seed, err := extractSeed(r)
	if err != nil {
		TextError(w, "seed: bad parameter", http.StatusBadRequest)
		return
	}

	rnd := rand.New(rand.NewSource(seed))
	body := randomBytes(totalBytes, rnd)

	w.Header().Set("Content-Length", strconv.Itoa(totalBytes))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// StreamRandomBytesHandle streams `n` random bytes generated with given `seed`, at given `chunk_size` per packet.
func StreamRandomBytesHandle(w http.ResponseWriter, r *http.Request) {
	// total bytes
	totalBytes, err := extractTotalBytes(r)
	if err != nil {
		TextError(w, "n: bad parameter", http.StatusBadRequest)
		return
	}

	seed, err := extractSeed(r)
	if err != nil {
		TextError(w, "seed: bad parameter", http.StatusBadRequest)
		return
	}

	var chunkSize int
	if chunkSizeParam := r.URL.Query().Get("chunk_size"); chunkSizeParam != "" {
		chunkSize, _ = strconv.Atoi(chunkSizeParam)
		chunkSize = max(1, chunkSize)
	} else {
		chunkSize = 1024 * 10
	}

	rnd := rand.New(rand.NewSource(seed))

	flusher := w.(http.Flusher)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// chunk_size can be bigger than totalBytes -- we will just send totalBytes in one chunk
	// on each iteration we choose from chunkSize or remaining bytes

	remBytes := totalBytes

	for {
		chunked := min(remBytes, chunkSize)
		body := randomBytes(chunked, rnd)
		w.Write(body)
		flusher.Flush()

		remBytes = remBytes - chunked
		if remBytes == 0 {
			break
		}
	}
}

func randomBytes(totalBytes int, rnd *rand.Rand) []byte {
	var body []byte

	for i := 0; i < totalBytes; i++ {
		body = append(body, byte(rnd.Intn(256)))
	}

	return body
}

func extractTotalBytes(r *http.Request) (totalBytes int, err error) {
	n := chi.URLParam(r, "n")
	totalBytes, err = strconv.Atoi(n)
	if err != nil {
		return
	}
	totalBytes = min(totalBytes, 100*1024)
	return
}

func extractSeed(r *http.Request) (seed int64, err error) {
	seedParam := r.URL.Query().Get("seed")
	if seedParam == "" {
		seed = time.Now().UnixNano()
	} else {
		seed, err = strconv.ParseInt(seedParam, 10, 64)
	}
	return
}
