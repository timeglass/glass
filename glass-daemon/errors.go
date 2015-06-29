package main

import (
	"log"
	"runtime"
	"strings"

	"github.com/timeglass/glass/_vendor/github.com/kardianos/service"
)

// give some more extensive information about the nature
// of a service control error
func ReportServiceControlErrors(err error) {
	if strings.Contains(err.Error(), "Unknown action") {
		log.Fatalf("Given action is invalid, only supports: %q", service.ControlAction)
	}

	if runtime.GOOS == "windows" {
		if strings.Contains(err.Error(), "Access is denied") {
			log.Fatalf("Don't have permission for service controls, make sure you run this as the administrator")
		}
	}

	log.Fatalf("Failed to handle service control: %s", err)
}
