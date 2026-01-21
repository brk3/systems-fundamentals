package server

import (
	"fmt"
	"net"
	"strconv"
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
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\nContent-Type: text/plain\nContent-Length: 13\n\nHello World!\n")
	conn.Close()
}

func (s *Server) Close() error {
	return s.listener.Close()
}
