package httpbulb

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
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

// RelativeRedirectHandle is a http handler that makes 302/3XX redirects `n` times.
// `Location` header is a relative URL.
func RelativeRedirectHandle(w http.ResponseWriter, r *http.Request) {
	redirectHandle(w, r, false)
}

// AbsoluteRedirectHandle is a http handler that makes 302/3XX redirects `n` times.
// `Location` header is an absolute URL.
func AbsoluteRedirectHandle(w http.ResponseWriter, r *http.Request) {
	redirectHandle(w, r, true)
}

// RedirectHandle is a http handler that makes 302/3XX redirects `n` times.
// `n` is a number in the URL path, if `n` is 1, it will redirect to `/get`.
// if `absolute` query param is true, `Location` header will be an absolute URL.
func RedirectHandle(w http.ResponseWriter, r *http.Request) {
	absolute := r.URL.Query().Get("absolute") == "true"
	redirectHandle(w, r, absolute)
}

func redirectHandle(w http.ResponseWriter, r *http.Request, absolute bool) {

	nParam := chi.URLParam(r, "n")

	n, err := strconv.Atoi(nParam)
	// actually this case is impossible, bad request is handled by chi router,
	// and n can be matched only as a number, chi will return 404 if it's not a number
	if err != nil {
		TextError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if n < 1 {
		TextError(w, "n must be greater than 0", http.StatusBadRequest)
		return
	}

	var locURL string
	var p string
	if n == 1 {
		p = "/get"
	}

	if absolute {
		if p == "" {
			p = fmt.Sprintf("/absolute-redirect/%d", n-1)
		}
		u := url.URL{Scheme: getURLScheme(r), Host: r.Host, Path: p}
		locURL = u.String()
	} else if p == "" {
		locURL = fmt.Sprintf("/relative-redirect/%d", n-1)
	} else {
		locURL = p
	}

	http.Redirect(w, r, locURL, http.StatusFound)

}
