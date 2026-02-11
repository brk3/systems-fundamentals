package server

import (
	"fmt"
	"log"
	"net"

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
			log.Printf("error accepting connection: %v", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("error reading request: %v", err)
	}

	response := &response.Writer{
		Writer:      conn,
		WriterState: response.WriterStateStatusLine,
	}
	s.handler(response, req)
}

func (s *Server) Close() error {
	return s.listener.Close()
}
