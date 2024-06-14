package httpbulb

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// BasicAuthHandle prompts the user for authorization using HTTP Basic Auth.
// It returns 401 if not authorized.
func BasicAuthHandle(w http.ResponseWriter, r *http.Request) {
	basicAuthHandle(w, r, http.StatusUnauthorized)
}

// HiddenBasicAuthHandle prompts the user for authorization using HTTP Basic Auth.
// It returns 404 if not authorized.
func HiddenBasicAuthHandle(w http.ResponseWriter, r *http.Request) {
	basicAuthHandle(w, r, http.StatusNotFound)
}

// BearerAuthHandle prompts the user for authorization using bearer authentication
func BearerAuthHandle(w http.ResponseWriter, r *http.Request) {
	authPrefix := "Bearer "
	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, authPrefix) {
		w.Header().Set("WWW-Authenticate", `Bearer"`)
		JsonError(w, "", http.StatusUnauthorized)
		return
	}
	token := authorization[len(authPrefix):]
	writeJsonResponse(w, http.StatusOK, AuthResponse{Authenticated: true, Token: token})
}

func basicAuthHandle(w http.ResponseWriter, r *http.Request, errCode int) {
	userParam := chi.URLParam(r, "user")
	passwdParam := chi.URLParam(r, "passwd")

	user, passwd, ok := r.BasicAuth()

	authenticated := user == userParam && passwd == passwdParam

	if !ok || !authenticated {
		w.Header().Set("WWW-Authenticate", `Basic realm="httpbulb"`)
		JsonError(w, "", errCode)
		return
	}

	writeJsonResponse(w, http.StatusOK, AuthResponse{Authenticated: true, User: user})

}
