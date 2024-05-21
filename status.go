package httpbulb

import (
	"encoding/json"
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"status_text": http.StatusText(statusCode)})
}
