package httpbulb

import "net/http"

// MethodsResponse is the response for the methods endpoint
type MethodsResponse struct {
	// Args is a map of query parameters
	Args map[string][]string `json:"args"`
	// Data is the raw body of the request
	Data string `json:"data"`
	// Files is a map of files sent in the request
	Files map[string][]string `json:"files"`
	// Form is a map of form values sent in the request
	Form map[string][]string `json:"form"`
	// Headers is a map of headers sent in the request
	Headers map[string][]string `json:"headers"`
	// JSON is the parsed JSON body of the request
	JSON interface{} `json:"json"`
	// Origin is the IP address of the requester
	Origin string `json:"origin"`
	// URL is the full URL of the request
	URL string `json:"url"`
	// Gzipped is true if the request was compressed with gzip
	Gzipped bool `json:"gzipped,omitempty"`
	// Brotli is true if the request was compressed with brotli
	Brotli bool `json:"brotli,omitempty"`
	// Deflated is true if the request was compressed with zlib
	Deflated bool `json:"deflated,omitempty"`
}

// StatusResponse is the response for the status endpoint
type StatusResponse struct {
	StatusText string `json:"status_text"`
}

// HeadersResponse is the response for the headers endpoint
type HeadersResponse struct {
	Headers http.Header `json:"headers"`
}

// IpResponse is the response for the ip endpoint
type IpResponse struct {
	Origin string `json:"origin"`
}

// UserAgentResponse is the response for the user-agent endpoint
type UserAgentResponse struct {
	UserAgent string `json:"user-agent"`
}

// AuthResponse is the response for the basic-auth endpoint
type AuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	User          string `json:"user"`
}
