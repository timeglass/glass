package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/daemon/watching"
	"github.com/timeglass/glass/model"
)

var Version = "0.0.0"
var Build = "gobuild"

var mbu = flag.Duration("mbu", time.Minute*6, "The minimal billable unit")
var bind = flag.String("bind", ":0", "Address to bind the Daemon to")
var force = flag.Bool("force", false, "Force start the Daemon")

func main() {
	flag.Parse()

	//by default timeout is four times the mbu
	//@todo make configurable
	timer := NewTimer(*mbu, 4*(*mbu))
	svr, err := NewServer(*bind, timer)
	if err != nil {
		log.Fatal(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err))
	}

	monitor, err := watching.NewMonitor(dir)
	if err != nil {
		log.Fatal(errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for directory '%s': {{err}}"), err))
	}

	//whenever _something_ happends in any directory of the project delay timeout
	timer.Wakeup = monitor.Events()
	go func() {
		for err := range monitor.Errors() {
			log.Printf("Monitor Error: %s", err)
		}
	}()

	err = monitor.Start()
	if err != nil {
		log.Fatal(errwrap.Wrapf(fmt.Sprintf("Failed to start monitor for directory '%s': {{err}}"), err))
	}

	m := model.New(dir)
	info, err := m.ReadDaemonInfo()
	if err != nil {
		log.Fatal(errwrap.Wrapf("Failed read Daemon info: {{err}}", err))
	}

	if info.Addr != "" && !*force {
		log.Fatal("It appears another Daemon is already running or a previous instance didn't shutdown properly, use -force to force start.")
	}

	info = model.NewDeamon(dir, svr.Addr())
	err = m.UpsertDaemonInfo(info)
	if err != nil {
		log.Fatal(errwrap.Wrapf("Failed write Daemon info: {{err}}", err))
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		svr.Stop()
	}()

	log.Printf("Listening on '%s'", svr.Addr())
	err = svr.Start()
	if err != nil && !strings.Contains(err.Error(), "closed network connection") {
		log.Fatal(err)
	}

	log.Printf("Writing information to database...")
	info.Addr = ""
	err = m.UpsertDaemonInfo(info)
	if err != nil {
		log.Fatal(errwrap.Wrapf("Failed write Daemon info: {{err}}", err))
	}

	log.Printf("Done")
}
