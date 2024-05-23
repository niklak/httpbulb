package httpbulb

import "net/http"

type MethodsResponse struct {
	Args    map[string][]string `json:"args"`
	Data    string              `json:"data"`
	Files   map[string][]string `json:"files"`
	Form    map[string][]string `json:"form"`
	Headers map[string][]string `json:"headers"`
	JSON    interface{}         `json:"json"`
	Origin  string              `json:"origin"`
	URL     string              `json:"url"`
}

type StatusResponse struct {
	StatusText string `json:"status_text"`
}

type HeadersResponse struct {
	Headers http.Header `json:"headers"`
}

type IpResponse struct {
	Origin string `json:"origin"`
}

type UserAgentResponse struct {
	UserAgent string `json:"user-agent"`
}
