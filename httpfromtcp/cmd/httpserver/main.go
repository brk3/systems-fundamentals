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

func writeHTML(w *response.Writer, statusCode response.StatusCode, body string) {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		_ = fmt.Errorf("error writing status line: %v", err)
	}

	defaultHeaders := response.GetDefaultHeaders(len(body))
	defaultHeaders["content-type"] = "text/html"
	err = w.WriteHeaders(defaultHeaders)
	if err != nil {
		_ = fmt.Errorf("error writing headers: %v", err)
	}

	_, err = w.WriteBody([]byte(body))
	if err != nil {
		_ = fmt.Errorf("error writing body: %v", err)
	}
}

func testHandler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`
		writeHTML(w, response.StatusBadRequest, body)
	case "/myproblem":
		body := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`
		writeHTML(w, response.StatusError, body)
	default:
		body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`
		writeHTML(w, response.StatusOk, body)
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
