package httpbulb

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

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

func parseRequestRange(rangeHeader string) (firstPos *int, lastPos *int) {
	if rangeHeader == "" {
		return
	}

	rangeHeader = strings.TrimSpace(rangeHeader)

	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return
	}

	value := strings.TrimPrefix(rangeHeader, "bytes=")

	components := strings.Split(value, "-")

	if len(components) > 0 {
		if left, err := strconv.Atoi(components[0]); err == nil {
			firstPos = &left
		}
	}

	if len(components) > 1 {
		if right, err := strconv.Atoi(components[1]); err == nil {
			lastPos = &right
		}
	}

	return
}

func getRequestRange(rangeHeader string, upperBound int) (firstPos int, lastPos int) {

	firstPosPtr, lastPosPtr := parseRequestRange(rangeHeader)

	if firstPosPtr == nil && lastPosPtr == nil {
		// Request full range
		firstPos = 0
		lastPos = upperBound - 1
	} else if firstPosPtr == nil {
		// Request the last X bytes
		firstPos = max(0, upperBound-*lastPosPtr)
		lastPos = upperBound - 1
	} else if lastPosPtr == nil {
		firstPos = *firstPosPtr
		lastPos = upperBound - 1
	} else {
		firstPos = *firstPosPtr
		lastPos = *lastPosPtr
	}

	return
}
