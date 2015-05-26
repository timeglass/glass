package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

var CheckVersionURL = "https://s3-eu-west-1.amazonaws.com/timeglass/version/VERSION?dversion=" + Version

type Server struct {
	timer *Timer
	httpb string

	listener          net.Listener
	mostRecentVersion string

	*http.Server
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Stopping Timer at: %s", s.timer.Time())
	defer s.Stop()
}

func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	s.timer.Start()
	fmt.Fprintf(w, "Started at: %s", s.timer.Time())
}

func (s *Server) pause(w http.ResponseWriter, r *http.Request) {
	s.timer.Stop()
	fmt.Fprintf(w, "Stopped at: %s", s.timer.Time())
}

func (s *Server) lap(w http.ResponseWriter, r *http.Request) {
	defer s.timer.Reset()
	s.Respond(w, map[string]interface{}{
		"Time": s.timer.Time().String(),
	})
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"CurrentVersion":    Version,
		"MostRecentVersion": s.mostRecentVersion,
		"Time":              s.timer.Time().String(),
	}

	//check version without delaying response
	go s.checkVersion()
	s.Respond(w, data)
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Daemon %s (%s)", Version, Build)
}

func NewServer(httpb string, timer *Timer) (*Server, error) {
	l, err := net.Listen("tcp", httpb)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create listener on '%s': {{err}}", httpb), err)
	}

	s := &Server{
		timer:             timer,
		httpb:             httpb,
		listener:          l,
		mostRecentVersion: Version,
		Server:            &http.Server{Handler: http.DefaultServeMux},
	}

	http.HandleFunc("/", version)
	http.HandleFunc("/timer.status", s.status)
	http.HandleFunc("/timer.pause", s.pause)
	http.HandleFunc("/timer.lap", s.lap)
	http.HandleFunc("/timer.start", s.start)
	http.HandleFunc("/timer.stop", s.stop)

	return s, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Stop() {
	s.listener.Close()
}

func (s *Server) Start() error {
	s.timer.Start()
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

func (s *Server) checkVersion() {
	resp, err := http.Get(CheckVersionURL)
	if err == nil {
		defer resp.Body.Close()
		buff := bytes.NewBuffer(nil)
		_, err = io.Copy(buff, resp.Body)
		if err != nil {
			log.Printf("Failed to read response body for version check: %s", err)
		} else {
			s.mostRecentVersion = strings.TrimSpace(buff.String())
		}
	}
}
