package httpbulb

import (
	"net/http"
	"strings"
)

func HeadersHandle(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, HeadersResponse{Headers: r.Header})
}

func IpHandle(w http.ResponseWriter, r *http.Request) {

	var ip string
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = strings.TrimSpace(strings.SplitN(forwardedFor, ",", 2)[0])
	} else {
		ip = r.RemoteAddr
	}
	writeJsonResponse(w, http.StatusOK, IpResponse{Origin: ip})
}

func UserAgentHandle(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, UserAgentResponse{UserAgent: r.UserAgent()})
}
