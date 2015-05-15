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

func version(c *echo.Context) *echo.HTTPError {
	return c.String(http.StatusOK, fmt.Sprintf("Daemon %s (%s)", Version, Build))
}

func NewServer(httpb string, timer *Timer) (*Server, error) {
	router := echo.New()
	router.Get("/", version)

	l, err := net.Listen("tcp", httpb)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create listener on '%s': {{err}}", httpb), err)
	}

	return &Server{
		timer:    timer,
		httpb:    httpb,
		router:   router,
		listener: l,
		Server:   &http.Server{Handler: router},
	}, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Start() error {
	err := s.timer.Start()
	if err != nil {
		return errwrap.Wrapf("Failted to start Timer: {{err}}", err)
	}

	return s.Server.Serve(s.listener)
}
