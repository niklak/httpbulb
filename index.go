package httpbulb

import "net/http"

// IndexHandle serves the index.html file
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	serveFileFS(w, r, assetsFS, "assets/index.html")
}

// StyleHandle serves the style.css file
func StyleHandle(w http.ResponseWriter, r *http.Request) {
	serveFileFS(w, r, assetsFS, "assets/style.css")
}
