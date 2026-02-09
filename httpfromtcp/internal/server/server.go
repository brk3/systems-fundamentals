package server

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	fmt.Println("Serve()")
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: ln,
		handler:  handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) listen() {
	fmt.Println("listen()")
	for {
		// Wait for a connection.
		conn, err := s.listener.Accept()
		if err != nil {
			panic(err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		_ = fmt.Errorf("error reading request: %v", err)
	}

	hBuf := &bytes.Buffer{}
	hErr := s.handler(hBuf, req)
	statusCode := response.StatusOk
	body := hBuf.String()
	if hErr != nil {
		statusCode = hErr.StatusCode
		body = hErr.Message
	}

	err = response.WriteStatusLine(conn, statusCode)
	if err != nil {
		_ = fmt.Errorf("error writing status line: %v", err)
	}

	h := response.GetDefaultHeaders(len(body))
	err = response.WriteHeaders(conn, h)
	if err != nil {
		_ = fmt.Errorf("error writing headers: %v", err)
	}

	if len(body) > 0 {
		fmt.Fprintf(conn, "\r\n")
		fmt.Fprint(conn, body)
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}
