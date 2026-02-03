package server

import (
	"fmt"
	"net"
	"strconv"

	"httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	fmt.Println("Serve()")
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	s := &Server{listener: ln}
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
	// TODO: this is breaking the connection prematurely
	// defer conn.Close()

	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		panic(err)
	}

	body := "Hello World!\n"
	h := response.GetDefaultHeaders(len(body))
	err = response.WriteHeaders(conn, h)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(conn, "\n")
	fmt.Fprint(conn, body)
}

func (s *Server) Close() error {
	return s.listener.Close()
}
