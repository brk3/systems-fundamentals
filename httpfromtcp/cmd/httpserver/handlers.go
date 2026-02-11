package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

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

func videoHandler(w *response.Writer, req *request.Request) {
	b, err := os.ReadFile("/Users/pbourke/sandbox/boot.dev/httpfromtcp/assets/vim.mp4")
	if err != nil {
		log.Printf("error reading local video file: %v", err)
		return
	}

	// write status line
	err = w.WriteStatusLine(response.StatusOk)
	if err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}

	// write headers
	h := response.GetDefaultHeaders(len(b))
	h.Set("Content-Type", "video/mp4")
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("error writing headers: %v", err)
		return
	}

	// write binary body
	_, err = w.WriteBody(b)
	if err != nil {
		log.Printf("error writing video to response: %v", err)
		return
	}
}

func httpBinHandler(w *response.Writer, req *request.Request) {
	// write status line
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

	// write headers
	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("error writing headers: %v", err)
	}

	// write chunked body
	buf := make([]byte, 1024)
	hash := sha256.New()
	dataLen := 0
	for {
		n, readErr := res.Body.Read(buf)
		fmt.Printf("XXX: read %d bytes from httpbin.com\n", n)
		// process any bytes we may have got before checking err
		if n > 0 {
			dataLen += n
			_, writeErr := w.WriteChunkedBody(buf[:n])
			if writeErr != nil {
				log.Printf("error writing chunk: %v", writeErr)
				return
			}
			_, err = hash.Write(buf[:n])
			if err != nil {
				log.Printf("error hashing chunk: %v", err)
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

	// write trailers
	trailers := headers.Headers{
		"X-Content-SHA256": fmt.Sprintf("%x", hash.Sum(nil)),
		"X-Content-Length": fmt.Sprintf("%d", dataLen),
	}
	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Printf("error writing trailers: %v", err)
		return
	}
}
