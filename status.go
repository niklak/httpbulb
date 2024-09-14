package httpbulb

import (
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// StatusCodesHandle returns status code or random status code if more than one are given.
// This handler does not handle status codes lesser than 200 or greater than 599.
func StatusCodeHandle(w http.ResponseWriter, r *http.Request) {
	var err error

	rawStatusCodes := chi.URLParam(r, "codes")
	rawStatusCodes, err = url.PathUnescape(rawStatusCodes)

	if err != nil {
		RenderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	parts := strings.Split(rawStatusCodes, ",")

	var codes []int

	for _, part := range parts {
		var code int
		code, err = strconv.Atoi(part)
		if err != nil {
			RenderError(w, err.Error(), http.StatusBadRequest)
			return
		}

		// skipping 1xx codes
		if code < 200 || code > 599 {
			RenderError(w, "status codes must be between 200 and 599", http.StatusBadRequest)
			return
		}
		codes = append(codes, code)
	}

	// verify that all codes are valid
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	statusCode := codes[rnd.Intn(len(codes))]
	statusText := http.StatusText(statusCode)
	if statusText == "" {
		statusText = "UNKNOWN"
	}

	RenderResponse(
		w, statusCode,
		StatusResponse{StatusText: statusText},
	)

}
