package httpbulb

import "net/http"

func RobotsHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	body := []byte(`
User-agent: *
Disallow: /deny
	`)
	w.Write(body)
}
