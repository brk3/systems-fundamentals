package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func testHandler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body := "Your problem is not my problem\n"

		err := w.WriteStatusLine(response.StatusBadRequest)
		if err != nil {
			_ = fmt.Errorf("error writing status line: %v", err)
		}

		defaultHeaders := response.GetDefaultHeaders(len(body))
		err = w.WriteHeaders(defaultHeaders)
		if err != nil {
			_ = fmt.Errorf("error writing headers: %v", err)
		}

		_, err = w.WriteBody([]byte(body))
		if err != nil {
			_ = fmt.Errorf("error writing body: %v", err)
		}
		// case "/myproblem":
		// 	return &server.HandlerError{
		// 		StatusCode: response.StatusError,
		// 		Message:    "Woopsie, my bad\n",
		// 	}
		// default:
		// 	w.Write([]byte("All good, frfr\n"))
	}
}

func main() {
	server, err := server.Serve(port, testHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
