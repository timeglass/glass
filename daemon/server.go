package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/labstack/echo"
)

var CheckVersionURL = "https://s3-eu-west-1.amazonaws.com/timeglass/version/VERSION?dversion=" + Version

type Server struct {
	timer             *Timer
	httpb             string
	router            *echo.Echo
	listener          net.Listener
	mostRecentVersion string

	*http.Server
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Stopping Timer at: %s", s.timer.Time())
	defer s.Stop()
}

func (s *Server) start(c *echo.Context) *echo.HTTPError {
	s.timer.Start()
	return c.String(http.StatusOK, fmt.Sprintf("Started at: %s", s.timer.Time()))
}

func (s *Server) pause(c *echo.Context) *echo.HTTPError {
	s.timer.Stop()
	return c.String(http.StatusOK, fmt.Sprintf("Stopped at: %s", s.timer.Time()))
}

func (s *Server) lap(c *echo.Context) *echo.HTTPError {
	defer s.timer.Reset()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"Time": s.timer.Time().String(),
	})
}

func (s *Server) status(c *echo.Context) *echo.HTTPError {
	data := map[string]interface{}{
		"CurrentVersion":    Version,
		"MostRecentVersion": s.mostRecentVersion,
		"Time":              s.timer.Time().String(),
	}

	//check version without delaying response
	go s.CheckVersion()

	return c.JSON(http.StatusOK, data)
}

func version(c *echo.Context) *echo.HTTPError {
	return c.String(http.StatusOK, fmt.Sprintf("Daemon %s (%s)", Version, Build))
}

func NewServer(httpb string, timer *Timer) (*Server, error) {
	router := echo.New()
	l, err := net.Listen("tcp", httpb)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create listener on '%s': {{err}}", httpb), err)
	}

	s := &Server{
		timer:             timer,
		httpb:             httpb,
		router:            router,
		listener:          l,
		mostRecentVersion: Version,
		Server:            &http.Server{Handler: router},
	}

	router.Get("/", version)
	router.Get("/timer.status", s.status)
	router.Get("/timer.pause", s.pause)
	router.Get("/timer.lap", s.lap)
	router.Get("/timer.start", s.start)
	router.Get("/timer.stop", s.stop)

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

func (s *Server) CheckVersion() {
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
