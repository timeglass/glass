package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/labstack/echo"
)

type Server struct {
	httpb    string
	router   *echo.Echo
	listener net.Listener

	*http.Server
}

//@todo show version
func info(c *echo.Context) *echo.HTTPError {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func NewServer(httpb string) (*Server, error) {
	router := echo.New()
	router.Get("/", info)

	l, err := net.Listen("tcp", httpb)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create listener on '%s': {{err}}", httpb), err)
	}

	return &Server{
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
	return s.Server.Serve(s.listener)
}
