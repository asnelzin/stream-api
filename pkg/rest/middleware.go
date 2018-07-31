package rest

import (
	"log"
	"net/http"
	"strings"
)

var org = "asnelzin"

// AppInfo adds custom app-info to the response header
func AppInfo(app string, version string) func(http.Handler) http.Handler {
	f := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Org", org)
			w.Header().Set("App-Name", app)
			w.Header().Set("App-Version", version)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}

// Ping middleware response with pong to /ping. Stops chain if ping request detected
func Ping(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" && strings.HasSuffix(strings.ToLower(r.URL.Path), "/ping") {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("pong")); err != nil {
				log.Printf("[WARN] can't send pong, %s", err)
			}
			return
		}
		next.ServeHTTP(w, r)

	}
	return http.HandlerFunc(fn)
}
