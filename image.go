package httpbulb

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ImageHandle returns a simple image of the type specified in the URL path.
// The supported image types are svg, jpeg, png, and webp.
func ImageHandle(w http.ResponseWriter, r *http.Request) {

	// Get the image format from the URL path parameter
	imgFormat := chi.URLParam(r, "format")

	var imgPath string
	switch imgFormat {
	case "svg", "jpeg", "png", "webp":
		imgPath = fmt.Sprintf("assets/images/im.%s", imgFormat)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Serve the image file
	http.ServeFileFS(w, r, assetsFS, imgPath)
}

// ImageAcceptHandle returns an image based on the client's Accept header.
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
	http.ServeFileFS(w, r, assetsFS, imgPath)

}
