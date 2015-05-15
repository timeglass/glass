package main

import (
	"log"
)

func main() {
	bind := ":0"

	svr, err := NewServer(bind)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on '%s'", svr.Addr())
	log.Fatal(svr.Start())
}
