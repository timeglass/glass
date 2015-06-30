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

func (s *Server) timersDelete(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.Respond(w, err)
		return
	}

	if confs, ok := r.Form["conf"]; !ok {
		s.Respond(w, fmt.Errorf("conf parameter is mandatory"))
		return
	} else {
		for _, confPath := range confs {
			err := s.keeper.Remove(confPath)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed to remove timer: {{err}}", err))
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) timersCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.Respond(w, err)
		return
	}

	if confs, ok := r.Form["conf"]; !ok {
		s.Respond(w, fmt.Errorf("conf parameter is mandatory"))
		return
	} else {
		for _, confPath := range confs {
			t, err := NewTimer(confPath)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed to create new timer: {{err}}", err))
				return
			}

			err = s.keeper.Add(t)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed to add new timer to keeper: {{err}}", err))
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"build":   Build,
		"version": Version,
		"keeper":  s.keeper,
	}

	s.Respond(w, data)
}

func NewServer(httpb string, keeper *Keeper) (*Server, error) {
	l, err := net.Listen("tcp", httpb)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create listener on '%s': {{err}}", httpb), err)
	}

	mux := http.NewServeMux()
	s := &Server{
		keeper:   keeper,
		httpb:    httpb,
		listener: l,

		Server: &http.Server{Handler: mux},
	}

	mux.HandleFunc("/api/", s.api)
	mux.HandleFunc("/api/timers.create", s.timersCreate)
	mux.HandleFunc("/api/timers.delete", s.timersDelete)
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
	var err error

	enc := json.NewEncoder(w)
	if derr, ok := data.(error); !ok {
		err = enc.Encode(data)
	} else {
		err = derr
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
	}
}
