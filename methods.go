package httpbulb

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func newMethodResponse(r *http.Request) (response MethodsResponse, err error) {

	var body []byte
	response = MethodsResponse{
		Args:    r.URL.Query(),
		Headers: getRequestHeader(r),
		Origin:  getIP(r),
		URL:     getAbsoluteURL(r),
	}

	ct, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";")

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		// do not read the body for GET and DELETE or any other requests
		return
	}

	switch ct {
	case "multipart/form-data":
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
	case "application/x-www-form-urlencoded":
		if err = r.ParseForm(); err != nil {
			return
		}
		response.Form = r.Form

	case "application/json":
		if err = json.NewDecoder(r.Body).Decode(&response.JSON); err != nil {
			return
		}
	default:
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return
		}
		response.Data = string(body)
	}

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
