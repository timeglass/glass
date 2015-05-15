package main

import (
	"flag"
	"log"
)

var Version = "0.0.0"
var Build = "gobuild"

var bind = flag.String("bind", ":0", "Address to bind the Daemon to")

func main() {
	flag.Parse()

	svr, err := NewServer(*bind)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on '%s'", svr.Addr())
	log.Fatal(svr.Start())
}
