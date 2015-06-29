package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"time"

	// "github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
	"github.com/timeglass/glass/_vendor/github.com/kardianos/service"
)

var Version = "0.0.0"
var Build = "gobuild"

type daemon struct{}

func (p *daemon) Start(s service.Service) error { go p.run(); return nil }
func (p *daemon) Stop(s service.Service) error  { return nil }
func (p *daemon) run() error {

	for {
		<-time.After(time.Second)
		log.Println("hello...")
	}

	return nil
}

func main() {
	flag.Parse()

	//setup logging to a file
	l, err := NewLogger(os.Stderr)
	if err != nil {
		log.Fatalf("Failed to create logger: %s", err)
	}

	log.SetOutput(l)
	defer l.Close()

	//initialize service
	conf := &service.Config{
		Name:        "timeglass",
		DisplayName: "Timeglass",
		Description: "Automated time tracking daemon that monitors file changes",
		Option:      map[string]interface{}{},
	}

	if runtime.GOOS == "darwin" {
		conf.Option["UserService"] = true
	}

	s, err := service.New(&daemon{}, conf)
	if err != nil {
		log.Fatal(err)
	}

	//handle service controls
	if len(flag.Args()) > 0 {
		err = service.Control(s, flag.Args()[0])
		if err != nil {
			ReportServiceControlErrors(err)
		}
		return
	}

	//start daemon
	log.Printf("Daemon launched, writing logs to '%s'", l.Path())
	defer func() {
		log.Printf("Daemon terminated\n\n")
	}()

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
