package httpbulb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Base64DecodeHandle decodes a base64 encoded string and returns the decoded value
func Base64DecodeHandle(w http.ResponseWriter, r *http.Request) {
	// Decode the base64 encoded string
	value := chi.URLParam(r, "value")

	decoded, err := base64.URLEncoding.DecodeString(value)

	if err != nil {
		RenderError(w, err.Error(), http.StatusBadRequest)
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
		RenderError(w, err.Error(), http.StatusBadRequest)
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
		RenderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	d = min(d, 10)

	delay := time.Duration(d) * time.Second

	// Delay for d milliseconds
	<-time.After(delay)

	resp, err := newMethodResponse(r)

	if err != nil {
		RenderError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RenderResponse(w, http.StatusOK, resp)

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

// UUIDHandle returns a new UUID version 4
func UUIDHandle(w http.ResponseWriter, r *http.Request) {
	RenderResponse(w, http.StatusOK, &UUIDResponse{UUID: uuid.New().String()})
}

// DripHandle drips data over a duration after an optional initial delay
func DripHandle(w http.ResponseWriter, r *http.Request) {

	var delay time.Duration
	if delayParam := r.URL.Query().Get("delay"); delayParam != "" {
		d, _ := strconv.Atoi(delayParam)
		delay = time.Duration(d) * time.Second
	}

	var code int
	if codeParam := r.URL.Query().Get("code"); codeParam != "" {
		code, _ = strconv.Atoi(codeParam)
	}

	if code == 0 {
		code = http.StatusOK
	}

	if code < 200 || code > 599 {
		RenderError(w, "code: status code must be between 200 and 599", http.StatusBadRequest)
		return
	}

	var numBytes int
	if numBytesParam := r.URL.Query().Get("numbytes"); numBytesParam != "" {
		numBytes, _ = strconv.Atoi(numBytesParam)
	}

	if numBytes <= 0 {
		RenderError(w, "numbytes: number of bytes must be positive", http.StatusBadRequest)
		return
	}
	// set max limit to 10MB
	numBytes = min(numBytes, 1024*1024*10)

	var duration time.Duration
	if durationParam := r.URL.Query().Get("duration"); durationParam != "" {
		d, _ := strconv.Atoi(durationParam)
		duration = time.Duration(d) * time.Second

	} else {
		duration = time.Second * 2
	}

	<-time.After(delay)

	pause := duration / time.Duration(numBytes)

	flusher := w.(http.Flusher)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(code)
	flusher.Flush()
	for i := 0; i < numBytes; i++ {
		w.Write([]byte{'*'})
		flusher.Flush()
		<-time.After(pause)
	}

}

// LinkPageHandle generates a page containing n links to other pages which do the same.
func LinkPageHandle(w http.ResponseWriter, r *http.Request) {
	nParam := chi.URLParam(r, "n")
	n, err := strconv.Atoi(nParam)
	if err != nil {
		RenderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	n = min(max(1, n), 200)

	offsetParam := chi.URLParam(r, "offset")

	offset, err := strconv.Atoi(offsetParam)
	if err != nil {
		RenderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "<html><head><title>Links</title></head><body>")

	for i := 0; i < n; i++ {
		if i == offset {
			fmt.Fprintf(w, `%d `, i)
		} else {
			fmt.Fprintf(w, `<a href='/links/%d/%d'>%d</a> `, n, i, i)
		}
	}
	fmt.Fprint(w, "</body></html>")
}

// LinksHandle redirects to first links page.
func LinksHandle(w http.ResponseWriter, r *http.Request) {
	nParam := chi.URLParam(r, "n")
	dst := fmt.Sprintf("/links/%s/0", nParam)
	http.Redirect(w, r, dst, http.StatusFound)
}

// RangeHandle streams n random bytes generated with given seed, at given chunk size per packet
func RangeHandle(w http.ResponseWriter, r *http.Request) {
	var err error
	var numBytes int
	if numBytes, err = strconv.Atoi(chi.URLParam(r, "numbytes")); err != nil {
		TextError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if numBytes < 0 || numBytes > (100*1024) {
		w.Header().Set("ETag", fmt.Sprintf("range%d", numBytes))
		w.Header().Set("Accept-Ranges", "bytes")
		TextError(w, "number of bytes must be in the range (0, 102400]", http.StatusNotFound)
		return
	}

	var chunkSize int
	if chunkSizeParam := r.URL.Query().Get("chunk_size"); chunkSizeParam != "" {
		chunkSize, _ = strconv.Atoi(chunkSizeParam)
		chunkSize = max(1, chunkSize)
	} else {
		chunkSize = 1024 * 10
	}

	var duration time.Duration
	if durationParam := r.URL.Query().Get("duration"); durationParam != "" {
		d, _ := strconv.Atoi(durationParam)
		duration = time.Duration(d) * time.Second
	}

	pausePerByte := duration / time.Duration(numBytes)

	firstBytePos, lastBytePos := getRequestRange(r.Header.Get("Range"), numBytes)

	if firstBytePos > lastBytePos ||
		firstBytePos > numBytes ||
		lastBytePos > numBytes {

		w.Header().Set("ETag", fmt.Sprintf("range%d", numBytes))
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", numBytes))
		w.Header().Set("Content-Length", "0")
		TextError(w, "", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	rangeLength := (lastBytePos + 1) - firstBytePos
	contentRange := fmt.Sprintf("bytes %d-%d/%d", firstBytePos, lastBytePos, numBytes)

	var statusCode int

	if firstBytePos == 0 && lastBytePos == numBytes-1 {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusPartialContent
	}

	flusher := w.(http.Flusher)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("ETag", fmt.Sprintf("range%d", numBytes))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.Itoa(rangeLength))
	w.Header().Set("Content-Range", contentRange)
	w.WriteHeader(statusCode)

	flusher.Flush()
	chunk := make([]byte, 0)
	for i := firstBytePos; i < lastBytePos+1; i++ {

		chunk = append(chunk, byte(97+(i%26)))
		if len(chunk) == chunkSize {
			w.Write(chunk)
			flusher.Flush()
			<-time.After(pausePerByte * time.Duration(chunkSize))
			chunk = make([]byte, 0)
		}
	}
	if len(chunk) > 0 {
		<-time.After(pausePerByte * time.Duration(len(chunk)))
		w.Write(chunk)
		flusher.Flush()

	}
}
