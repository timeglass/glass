package main

import (
	"flag"
	"log"
	"time"
)

var Version = "0.0.0"
var Build = "gobuild"

var mbu = flag.Duration("mbu", time.Minute*6, "The minimal billable unit")
var bind = flag.String("bind", ":0", "Address to bind the Daemon to")

func main() {
	flag.Parse()

	timer := NewTimer(*mbu)
	svr, err := NewServer(*bind, timer)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on '%s'", svr.Addr())
	log.Fatal(svr.Start())
}
