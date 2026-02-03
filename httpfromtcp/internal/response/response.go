package response

import (
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

type StatusCode int

const HTTPVersion = "HTTP/1.1"

const (
	statusOk         StatusCode = 200
	statusBadRequest StatusCode = 400
	statusError      StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case statusOk:
		_, err := fmt.Fprintf(w, "%s %d %s\n", HTTPVersion, statusCode, "OK")
		return err
	case statusBadRequest:
		_, err := fmt.Fprintf(w, "%s %d %s\n", HTTPVersion, statusCode, "Bad Request")
		return err
	case statusError:
		_, err := fmt.Fprintf(w, "%s %d %s\n", HTTPVersion, statusCode, "Internal Server Error")
		return err
	default:
		_, err := fmt.Fprintf(w, "%s %d\n", HTTPVersion, statusCode)
		return err
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\n", key, val)
		if err != nil {
			return err
		}
	}
	return nil
}
