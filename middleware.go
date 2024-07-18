package httpbulb

import "net/http"

// Cors middleware to handle Cross Origin Resource Sharing (CORS).
func Cors(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if acrh, ok := r.Header["Access-Control-Request-Headers"]; ok {
				for _, v := range acrh {
					w.Header().Add("Access-Control-Allow-Headers", v)
				}
			}
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
