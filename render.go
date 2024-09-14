package httpbulb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RenderResponse is the default renderer function used by httpbulb.
// It renders JSON by default and it can be overridden on program's `init` function.
var RenderResponse func(http.ResponseWriter, int, interface{}) = renderJson

// RenderError renders the error message.
// It renders JSON by default and it can be overridden on program's `init` function.
var RenderError func(http.ResponseWriter, string, int) = JsonError

func renderJson(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

// TextError is a shortcut func for writing an error in `text/plain`
func TextError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintln(w, err)
}

// JsonError is a shortcut func for writing an error in `application/json`
func JsonError(w http.ResponseWriter, err string, code int) {

	if err == "" {
		err = http.StatusText(code)
	}
	renderJson(w, code, &ErrorResponse{Error: err})
}
