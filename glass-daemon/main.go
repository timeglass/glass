package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
	"github.com/timeglass/glass/_vendor/github.com/kardianos/service"

	"github.com/timeglass/glass/model"
	"github.com/timeglass/snow/monitor"
)

var Version = "0.0.0"
var Build = "gobuild"

var mbu = flag.Duration("mbu", time.Minute, "The minimal billable unit")
var bind = flag.String("bind", ":0", "Address to bind the Daemon to")
var force = flag.Bool("force", false, "Force start the Daemon")

var logger *Logger

func main() {
	flag.Parse()
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err))
	}

	logger, err = NewLogger(dir, os.Stderr)
	if err != nil {
		log.Fatal(errwrap.Wrapf("Failed to create logger: {{err}}", err))
	}

	//by default timeout is four times the mbu
	//@todo make configurable
	timer := NewTimer(*mbu, 4*(*mbu))
	svr, err := NewServer(*bind, timer)
	if err != nil {
		logger.Fatal(err)
	}

	//check version without delaying start times
	go svr.checkVersion()

	m := model.New(dir)
	monitor, err := monitor.New(dir, monitor.Recursive, time.Millisecond*50)
	if err != nil {
		logger.Fatal(errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for directory '%s': {{err}}"), err))
	}

	go func() {
		for err := range monitor.Errors() {
			logger.Printf("Monitor Error: %s", err)
		}
	}()

	//whenever _something_ happens in any directory, wakeup the timer
	timer.Wakeup, err = monitor.Start()
	if err != nil {
		logger.Fatal(errwrap.Wrapf(fmt.Sprintf("Failed to start monitor for directory '%s': {{err}}"), err))
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		svr.Stop(nil)
	}()

	//
	// Service Setup
	//

	sConfig := &service.Config{
		Name:             "com.timeglass." + m.DirHash(),
		Arguments:        []string{"-force"},
		DisplayName:      fmt.Sprintf("Glass Timer for %s", dir),
		Description:      "Automated time tracking for Git repositories.",
		WorkingDirectory: dir,
		Option: map[string]interface{}{
			"UserService": true,
			"RunAtLoad":   true,
		},
	}

	s, err := service.New(svr, sConfig)
	if err != nil {
		logger.Fatal(err)
	}

	if len(flag.Args()) > 0 {
		err = service.Control(s, flag.Args()[0])
		if err != nil {
			logger.Fatal(err)
		}
		return
	}

	//
	// Model Access
	//

	info, err := m.ReadDaemonInfo()
	if err != nil {
		logger.Fatal(errwrap.Wrapf("Failed read Daemon info: {{err}}", err))
	}

	if info.Addr != "" && !*force {
		logger.Fatal("It appears another Daemon is already running or a previous instance didn't shutdown properly, use -force to force start.")
	}

	info = model.NewDeamon(dir, svr.Addr())
	err = m.UpsertDaemonInfo(info)
	if err != nil {
		logger.Fatal(errwrap.Wrapf("Failed write Daemon info: {{err}}", err))
	}

	//
	// Start Service
	//

	logger.Printf("Listening on '%s'", svr.Addr())
	err = s.Run()
	if err != nil && !strings.Contains(err.Error(), "closed network connection") {
		logger.Fatal(err)
	}

	logger.Printf("Writing information to database...")
	info.Addr = ""
	err = m.UpsertDaemonInfo(info)
	if err != nil {
		logger.Fatal(errwrap.Wrapf("Failed write Daemon info: {{err}}", err))
	}

	logger.Printf("Done")
}
