package httpbulb

import (
	"net/http"
)

// HeadersHandle returns only the request headers. Check `HeadersResponse`.
func HeadersHandle(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, HeadersResponse{Headers: getRequestHeader(r)})
}

// IpHandle returns only the IP address. Check `IpResponse`.
func IpHandle(w http.ResponseWriter, r *http.Request) {

	writeJsonResponse(w, http.StatusOK, IpResponse{Origin: getIP(r)})
}

// UserAgentHandle returns only the user agent. Check `UserAgentResponse`.
func UserAgentHandle(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, UserAgentResponse{UserAgent: r.UserAgent()})
}
