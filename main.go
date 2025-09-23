package main

import (
	"net/http"
	"os"

	"github.com/charmbracelet/log"
)

var Mnemo *MnemoServer
var Database *AuthDatabase

func main() {
	log.SetDefault(log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
		ReportCaller:    true,
	}))

	Mnemo = CreateMnemoServer(":8080")

	setup_handler()

	Mnemo.StartServer()
}

func setup_handler() {
	Mnemo.RegisterHandler("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to Mnemo!\n"))
	})
}
