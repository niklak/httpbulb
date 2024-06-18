package httpbulb

import (
	"net/http"
)

func HeadersHandle(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, HeadersResponse{Headers: getRequestHeader(r)})
}

func IpHandle(w http.ResponseWriter, r *http.Request) {

	writeJsonResponse(w, http.StatusOK, IpResponse{Origin: getIP(r)})
}

func UserAgentHandle(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, UserAgentResponse{UserAgent: r.UserAgent()})
}
