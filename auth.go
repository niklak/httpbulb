package httpbulb

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func BasicAuthHandle(w http.ResponseWriter, r *http.Request) {

	userParam := chi.URLParam(r, "user")
	passwdParam := chi.URLParam(r, "passwd")

	user, passwd, ok := r.BasicAuth()

	authenticated := user == userParam && passwd == passwdParam

	if !ok || !authenticated {
		w.Header().Set("WWW-Authenticate", `Basic realm="Fake Realm"`)
		JsonError(w, "", http.StatusUnauthorized)
		return
	}

	writeJsonResponse(w, http.StatusOK, AuthResponse{Authenticated: true, User: user})

}
