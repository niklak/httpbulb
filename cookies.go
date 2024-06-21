package httpbulb

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CookiesHandle returns the cookies list, sent by the client
func CookiesHandle(w http.ResponseWriter, r *http.Request) {
	cookies := r.Cookies()
	respCookies := make(map[string][]string)

	for _, cookie := range cookies {
		if _, ok := respCookies[cookie.Name]; !ok {
			respCookies[cookie.Name] = []string{}
		}
		respCookies[cookie.Name] = append(respCookies[cookie.Name], cookie.Value)
	}

	resp := CookiesResponse{Cookies: respCookies}
	writeJsonResponse(w, http.StatusOK, resp)
}

// SetCookiesHandle sets the cookies passed from the query parameters,
// then redirects to /cookies
func SetCookiesHandle(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	for k, vv := range params {
		for _, v := range vv {
			http.SetCookie(w, &http.Cookie{
				Name:     k,
				Value:    v,
				HttpOnly: true,
				Path:     "/",
			})
		}
	}

	http.Redirect(w, r, "/cookies", http.StatusFound)

}

// SetCookieHandle sets a cookie with the name and value passed in the URL path,
// then redirects to /cookies
func SetCookieHandle(w http.ResponseWriter, r *http.Request) {

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/cookies", http.StatusFound)

}

// DeleteCookiesHandle deletes the cookies passed in the query parameters,
// then redirects to /cookies
func DeleteCookiesHandle(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

	for param := range params {
		http.SetCookie(w, &http.Cookie{
			Name:   param,
			MaxAge: -1,
			Path:   "/",
		})
	}

	http.Redirect(w, r, "/cookies", http.StatusFound)

}

// CookiesListResponse returns a **list** with request cookies
func CookiesListHandle(w http.ResponseWriter, r *http.Request) {

	resp := CookiesListResponse{Cookies: r.Cookies()}
	writeJsonResponse(w, http.StatusOK, resp)
}
