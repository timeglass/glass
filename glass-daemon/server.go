package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	// "io"
	"log"
	"net"
	"net/http"
	// "strings"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Server struct {
	keeper   *Keeper
	httpb    string
	listener net.Listener

	*http.Server
}

func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"build":       Build,
		"version":     Version,
		"time_keeper": s.keeper.Data(),
	}

	s.Respond(w, data)
}

func NewServer(httpb string, keeper *Keeper) (*Server, error) {
	l, err := net.Listen("tcp", httpb)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create listener on '%s': {{err}}", httpb), err)
	}

	s := &Server{
		keeper:   keeper,
		httpb:    httpb,
		listener: l,

		Server: &http.Server{Handler: http.DefaultServeMux},
	}

	http.HandleFunc("/api/", s.api)

	return s, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Stop() error {
	return s.listener.Close()
}

func (s *Server) Start() error {
	log.Printf("Started server on %s", s.Addr())
	defer func() {
		log.Printf("Stopped server on %s", s.Addr())
	}()

	return s.Server.Serve(s.listener)
}

func (s *Server) Respond(w http.ResponseWriter, data interface{}) {
	enc := json.NewEncoder(w)
	err := enc.Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
	}
}
