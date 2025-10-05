package networking

import (
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

func (s *MnemoServer) GetAddress() string {
	return s.address
}

// func (s *MnemoServer) RegisterSessionValidatedHandler(pattern string, fn http.HandlerFunc) {
// 	wrapped_function := s.sessionMiddlewareHandler(http.HandlerFunc(fn))
// 	s.mux.Handle(pattern, wrapped_function)
// }

func (s *MnemoServer) RegisterHandler(pattern string, fn http.HandlerFunc) {
	s.mux.HandleFunc(pattern, fn)
}

func (s *MnemoServer) StartServer() error {
	log.Infof("Starting server on %v", s.address)
	s.server = &http.Server{
		Addr:    s.address,
		Handler: s.LogMiddlewareHandler(s.mux),
	}
	s.running = true
	return s.server.ListenAndServe()
}

// Middleware
