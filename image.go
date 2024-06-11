package httpbulb

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func ImageHandle(w http.ResponseWriter, r *http.Request) {

	imgFormat := chi.URLParam(r, "format")

	var imgPath string

	switch imgFormat {
	case "svg":
		imgPath = "assets/images/im.svg"
	case "jpeg":
		imgPath = "assets/images/im.jpeg"
	case "png":
		imgPath = "assets/images/im.png"
	case "webp":
		imgPath = "assets/images/im.webp"
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, imgPath)
}

// ImageAcceptHandle returns a simple image of the type suggest by the Accept header
func ImageAcceptHandle(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("Accept")
	var imgPath string
	if strings.Contains(accept, "image/webp") {
		imgPath = "assets/images/im.webp"
	} else if strings.Contains(accept, "image/svg+xml") {
		imgPath = "assets/images/im.svg"
	} else if strings.Contains(accept, "image/jpeg") {
		imgPath = "assets/images/im.jpeg"
	} else if strings.Contains(accept, "image/png") ||
		strings.Contains(accept, "image/*") {
		imgPath = "assets/images/im.png"
	} else {
		JsonError(
			w,
			"Client did not request a supported media type.",
			http.StatusNotAcceptable,
		)
		return
	}
	http.ServeFile(w, r, imgPath)

}
