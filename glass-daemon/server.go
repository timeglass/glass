package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/timer"
)

var CheckVersionURL = "https://s3-eu-west-1.amazonaws.com/timeglass/version/VERSION?dversion=" + Version

type Server struct {
	keeper            *timer.Keeper
	httpb             string
	listener          net.Listener
	mostRecentVersion string

	*http.Server
}

func (s *Server) timersDelete(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.Respond(w, err)
		return
	}

	if dirs, ok := r.Form["dir"]; !ok {
		s.Respond(w, fmt.Errorf("dir parameter is mandatory"))
		return
	} else {
		for _, dir := range dirs {
			err := s.keeper.Discard(dir)
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

	if dirs, ok := r.Form["dir"]; !ok {
		s.Respond(w, fmt.Errorf("dir parameter is mandatory"))
		return
	} else {
		for _, dir := range dirs {
			err = s.keeper.Measure(dir)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed to add new timer to keeper: {{err}}", err))
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) timersPause(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.Respond(w, err)
		return
	}

	if dirs, ok := r.Form["dir"]; !ok {
		s.Respond(w, fmt.Errorf("dir parameter is mandatory"))
		return
	} else {
		for _, dir := range dirs {
			t, err := s.keeper.Inspect(dir)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed get timer: {{err}}", err))
				return
			}

			t.Pause()
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) timersReset(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.Respond(w, err)
		return
	}

	if dirs, ok := r.Form["dir"]; !ok {
		s.Respond(w, fmt.Errorf("dir parameter is mandatory"))
		return
	} else {
		for _, dir := range dirs {
			t, err := s.keeper.Inspect(dir)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed get timer: {{err}}", err))
				return
			}

			staged, err := strconv.ParseBool(r.Form.Get("staged"))
			if err != nil {
				staged = false
			}

			t.Reset(staged)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) timersInfo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.Respond(w, err)
		return
	}

	timers := []*timer.Timer{}
	if dirs, ok := r.Form["dir"]; !ok {
		s.Respond(w, fmt.Errorf("dir parameter is mandatory"))
		return
	} else {
		for _, dir := range dirs {
			t, err := s.keeper.Inspect(dir)
			if err != nil {
				s.Respond(w, errwrap.Wrapf("Failed to get timer: {{err}}", err))
				return
			}

			timers = append(timers, t)
		}
	}

	s.Respond(w, timers)
}

func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"build":          Build,
		"version":        Version,
		"newest_version": s.mostRecentVersion,
		"keeper":         s.keeper,
	}

	go s.checkVersion()
	s.Respond(w, data)
}

func NewServer(httpb string, keeper *timer.Keeper) (*Server, error) {
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
	mux.HandleFunc("/api/timers.pause", s.timersPause)
	mux.HandleFunc("/api/timers.delete", s.timersDelete)
	mux.HandleFunc("/api/timers.reset", s.timersReset)
	mux.HandleFunc("/api/timers.info", s.timersInfo)
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
