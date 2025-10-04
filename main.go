package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/mnemosynefs/mnemo/internal/services"
)

func main() {
	log.SetDefault(log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
		ReportCaller:    true,
	}))

	services.CreateServices(":8080", "./auth.json")
}
