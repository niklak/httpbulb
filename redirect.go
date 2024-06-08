package httpbulb

import (
	"net/http"
	"strconv"
)

// RedirectToHandle is a http handler that makes 302/3XX redirects to the given URL
func RedirectToHandle(w http.ResponseWriter, r *http.Request) {
	var err error
	var dstURL string
	var rawStatusCode string
	switch r.Method {
	case http.MethodGet, http.MethodDelete:
		dstURL = r.URL.Query().Get("url")
		rawStatusCode = r.URL.Query().Get("status")
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			http.Error(w,
				"content-type 'application/x-www-form-urlencoded' is expected",
				http.StatusBadRequest,
			)
			return
		}
		if err = r.ParseForm(); err != nil {
			http.Error(w, "unable to parse form", http.StatusBadRequest)
			return
		}
		dstURL = r.Form.Get("url")

		rawStatusCode = r.Form.Get("status")
	}

	if dstURL == "" {
		http.Error(w, "'url' is required", http.StatusBadRequest)
		return
	}

	var statusCode int

	if rawStatusCode != "" {
		statusCode, err = strconv.Atoi(rawStatusCode)
		if err != nil {
			http.Error(w, "invalid status code", http.StatusBadRequest)
			return
		}
	}

	var responseStatusCode int
	if statusCode >= 300 && statusCode < 400 {
		responseStatusCode = statusCode
	} else {
		responseStatusCode = http.StatusFound
	}
	http.Redirect(w, r, dstURL, responseStatusCode)
}
