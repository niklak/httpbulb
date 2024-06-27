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
	case "svg", "jpeg", "png", "webp", "avif":
		imgPath = fmt.Sprintf("assets/images/im.%s", imgFormat)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Serve the image file
	serveFileFS(w, r, assetsFS, imgPath)
}

// ImageAcceptHandle returns an image based on the client's Accept header.
func ImageAcceptHandle(w http.ResponseWriter, r *http.Request) {

	//TODO: consider handling ;q= for weight or drop this idea
	//TODO: what about */*
	accept := r.Header.Get("Accept")
	cut := "image/"
	var found bool
	if _, accept, found = strings.Cut(accept, cut); !found {
		JsonError(w, "Client did not request a supported media type.", http.StatusNotAcceptable)
		return
	}

	if imgEndPos := strings.IndexAny(accept, ",;"); imgEndPos > 0 {
		accept = accept[:imgEndPos]
	}

	var imgPath string

	switch accept {
	case "jpeg", "png", "webp", "avif":
		imgPath = fmt.Sprintf("assets/images/im.%s", accept)
	case "svg+xml":
		imgPath = "assets/images/im.svg"
	case "*":
		imgPath = "assets/images/im.png"
	default:
		JsonError(w, "Client did not request a supported media type.", http.StatusNotAcceptable)
		return
	}
	serveFileFS(w, r, assetsFS, imgPath)

}
