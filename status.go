package httpbulb

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func statusCodeHandle(w http.ResponseWriter, r *http.Request) {
	var err error

	rawStatusCode := chi.URLParam(r, "code")

	statusCode, err := strconv.Atoi(rawStatusCode)

	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJsonResponse(
		w, statusCode,
		StatusResponse{StatusText: http.StatusText(statusCode)},
	)
}
