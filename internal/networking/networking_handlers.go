package networking

import (
	"net/http"

	"github.com/charmbracelet/log"
)

// Authentication handlers

//
// Middleware
//

func (s *MnemoServer) LogMiddlewareHandler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s - %s %s\n", r.RemoteAddr, r.Proto, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
