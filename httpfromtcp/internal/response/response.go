package response

import (
	"fmt"
	"io"

	"httpfromtcp/internal/headers"
)

const HTTPVersion = "HTTP/1.1"

type StatusCode int

const (
	StatusOk         StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusError      StatusCode = 500
)

type WriterState int

const (
	WriterStateStatusLine WriterState = iota
	WriterStateHeaders
	WriterStateBody
)

type Writer struct {
	Writer      io.Writer
	WriterState WriterState
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprint(contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState != WriterStateStatusLine {
		return fmt.Errorf("error: unexpected writerState %d when calling WriteStatusLine", w.WriterState)
	}
	var reason string
	switch statusCode {
	case StatusOk:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusError:
		reason = "Internal Server Error"
	default:
		reason = "Unknown"
	}

	_, err := fmt.Fprintf(w.Writer, "%s %d %s\r\n", HTTPVersion, statusCode, reason)
	if err == nil {
		w.WriterState = WriterStateHeaders
	}
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriterState != WriterStateHeaders {
		return fmt.Errorf("error: unexpected writerState %d when calling WriteHeaders", w.WriterState)
	}
	// TODO: merge with default
	for key, val := range headers {
		_, err := fmt.Fprintf(w.Writer, "%s: %s\n", key, val)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.Writer, "\r\n")
	if err != nil {
		return err
	}
	w.WriterState = WriterStateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterState != WriterStateBody {
		return 0, fmt.Errorf("error: unexpected writerState %d when calling WriteBody", w.WriterState)
	}
	n, err := w.Writer.Write(p)
	return n, err
}
