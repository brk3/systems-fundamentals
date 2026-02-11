package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func writeHTML(w *response.Writer, statusCode response.StatusCode, body string) {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("error writing status line: %v", err)
	}

	h := response.GetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("error writing headers: %v", err)
	}

	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Printf("error writing body: %v", err)
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

func httpBinHandler(w *response.Writer, req *request.Request) {
	s := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	res, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", s))
	if err != nil {
		log.Printf("error fetching from httpbin.com: %v", err)
		err = w.WriteStatusLine(response.StatusError)
		if err != nil {
			log.Printf("error writing status line: %v", err)
		}
		return
	}
	err = w.WriteStatusLine(response.StatusOk)
	if err != nil {
		log.Printf("error writing status line: %v", err)
	}
	defer res.Body.Close()

	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("error writing headers: %v", err)
	}

	buf := make([]byte, 1024)
	for {
		n, readErr := res.Body.Read(buf)
		fmt.Printf("XXX: read %d bytes from httpbin.com\n", n)
		// process any bytes we may have got before checking err
		if n > 0 {
			_, writeErr := w.WriteChunkedBody(buf[:n])
			if writeErr != nil {
				log.Printf("error writing chunk: %v", err)
				return
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			log.Printf("error reading data: %v", readErr)
			return
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("error writing closer chunk: %v", err)
		return
	}
	// if res.StatusCode > 299 {
	// 	log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	// }
}

func router(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		httpBinHandler(w, req)
	} else {
		testHandler(w, req)
	}
}

func main() {
	server, err := server.Serve(port, router)
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
