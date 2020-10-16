package httptestbench

import (
	"fmt"
	"net"
	"time"

	"github.com/valyala/fasthttp"
)

// Server is a fasthttp server for tests.
type Server struct {
	URL string // base URL of form http://ipaddr:port with no trailing slash

	srv *fasthttp.Server
}

// NewServer starts new fasthttp test server with free local port.
func NewServer(handler fasthttp.RequestHandler) *Server {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptestbench: failed to listen on a port: %v", err))
		}
	}

	s := &Server{
		URL: "http://" + l.Addr().String(),
	}

	s.srv = &fasthttp.Server{
		Handler:     handler,
		ReadTimeout: 10 * time.Second,
	}

	go func() {
		err := s.srv.Serve(l)
		if err != nil {
			panic(fmt.Sprintf("httptestbench: failed to serve: %v", err))
		}
	}()

	return s
}

// Close shutdowns test server.
func (s *Server) Close() {
	err := s.srv.Shutdown()
	if err != nil {
		panic(fmt.Sprintf("httptestbench: failed to shutdown test server: %v", err))
	}
}
