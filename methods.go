package httpbulb

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func newMethodResponse(r *http.Request) (response MethodsResponse, err error) {
	// TODO: add json support

	var body []byte
	response = MethodsResponse{
		Args:    r.URL.Query(),
		Headers: r.Header,
		Origin:  getIP(r),
		URL:     getAbsoluteURL(r),
	}
	ct := r.Header.Get("Content-Type")
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if strings.HasPrefix(ct, "multipart/form-data") {
			if err = r.ParseMultipartForm(64 << 20); err != nil {
				return
			}

			if r.MultipartForm != nil && r.MultipartForm.File != nil {
				files := make(map[string][]string)
				for k, f := range r.MultipartForm.File {
					for _, fileHeader := range f {
						var file multipart.File

						if file, err = fileHeader.Open(); err != nil {
							return
						}
						var fBody []byte

						if fBody, err = io.ReadAll(file); err != nil {
							return
						}
						files[k] = append(files[k], string(fBody))
					}
				}
				response.Files = files
				response.Form = r.Form

			}
		} else if ct == "application/x-www-form-urlencoded" {
			if err = r.ParseForm(); err != nil {
				return
			}
			response.Form = r.Form
		} else {
			body, err = io.ReadAll(r.Body)
			if err != nil {
				return
			}
			response.Data = string(body)
		}
	}

	response.JSON = nil
	return
}

// MethodsHandle is the basic handler for the methods endpoint (GET, POST, PUT, PATCH, DELETE)
func MethodsHandle(w http.ResponseWriter, r *http.Request) {

	var err error

	response, err := newMethodResponse(r)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, http.StatusOK, response)
}
