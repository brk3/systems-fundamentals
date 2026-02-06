package server

import (
	"fmt"
	"net"
	"strconv"

	"httpfromtcp/internal/request"
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
	defer conn.Close()

    // reader := bufio.NewReader(conn)
    // for {
    //     line, err := reader.ReadString('\n')
    //     if err != nil {
    //         break
    //     }
    //     // HTTP headers end with a blank line (\r\n)
    //     if line == "\r\n" || line == "\n" {
    //         break
    //     }
    // }
	_, err := request.RequestFromReader(conn)
	if err != nil {
		_ = fmt.Errorf("error reading request: %v", err)
	}

	err = response.WriteStatusLine(conn, 200)
	if err != nil {
		fmt.Println(err)
	}

	body := "Hello World!\n"
	h := response.GetDefaultHeaders(len(body))
	err = response.WriteHeaders(conn, h)
	if err != nil {
		_ = fmt.Errorf("error writing headers: %v", err)
	}

	fmt.Fprintf(conn, "\r\n")
	fmt.Fprint(conn, body)
}

func (s *Server) Close() error {
	return s.listener.Close()
}
