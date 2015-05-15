package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/labstack/echo"
)

type Server struct {
	timer    *Timer
	httpb    string
	router   *echo.Echo
	listener net.Listener

	*http.Server
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Stopping Timer at: %s", s.timer.Time())
	defer s.listener.Close()
}

func (s *Server) start(c *echo.Context) *echo.HTTPError {
	s.timer.Start()
	return c.String(http.StatusOK, fmt.Sprintf("Started at: %s", s.timer.Time()))
}

func (s *Server) pause(c *echo.Context) *echo.HTTPError {
	s.timer.Stop()
	return c.String(http.StatusOK, fmt.Sprintf("Stopped at: %s", s.timer.Time()))
}

func (s *Server) status(c *echo.Context) *echo.HTTPError {
	return c.String(http.StatusOK, fmt.Sprintf("Timer: %s", s.timer.Time()))
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
		timer:    timer,
		httpb:    httpb,
		router:   router,
		listener: l,
		Server:   &http.Server{Handler: router},
	}

	router.Get("/", version)
	router.Get("/timer.status", s.status)
	router.Get("/timer.pause", s.pause)
	router.Get("/timer.start", s.start)
	router.Get("/timer.stop", s.stop)

	return s, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Start() error {
	s.timer.Start()
	return s.Server.Serve(s.listener)
}
