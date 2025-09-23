package main

import (
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
)

type MnemoServer struct {
	address string
	mux     *http.ServeMux
	server  *http.Server

	running bool
}

func CreateMnemoServer(address string) *MnemoServer {
	mux := http.NewServeMux()

	mnemoServer := &MnemoServer{
		address: address,
		mux:     mux,
	}

	return mnemoServer
}

func (mnemo *MnemoServer) RegisterSessionValidatedHandler(pattern string, fn http.HandlerFunc) {
	wrapped_function := sessionMiddlewareHandler(http.HandlerFunc(fn))
	mnemo.mux.Handle(pattern, wrapped_function)
}

func (mnemo *MnemoServer) RegisterHandler(pattern string, fn http.HandlerFunc) {
	mnemo.mux.HandleFunc(pattern, fn)
}

func (mnemo *MnemoServer) StartServer() error {
	log.Infof("Starting server on %v", mnemo.address)
	mnemo.server = &http.Server{
		Addr:    mnemo.address,
		Handler: logMiddlewareHandler(mnemo.mux),
	}
	mnemo.running = true
	return mnemo.server.ListenAndServe()
}

// Middleware
func logMiddlewareHandler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s - %s %s\n", r.RemoteAddr, r.Proto, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func sessionMiddlewareHandler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		session_token := r.Header.Get("session_token")

		if session_token == "" {
			r.Header.Set("username", "")
			next.ServeHTTP(w, r)
			log.Info("Admitting guest account")
			return
		}

		username, err := Database.GetUserFromToken(session_token)

		if errors.Is(err, INVALID_SESSION) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid token"))
			log.Infof("Login attempt by invalid token: %v", session_token)
			return
		}

		r.Header.Set("username", username)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
