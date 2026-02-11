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
	h := headers.Headers{}
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")
	if contentLen > 0 {
		h.Set("content-length", fmt.Sprint(contentLen))
	}
	return h
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
	for key, val := range headers {
		_, err := fmt.Fprintf(w.Writer, "%s: %s\r\n", key, val)
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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriterState != WriterStateBody {
		return 0, fmt.Errorf("error: unexpected writerState %d when calling WriteChunkedBody", w.WriterState)
	}
	n, err := fmt.Fprintf(w.Writer, "%X\r\n", len(p))
	n1, err := fmt.Fprintf(w.Writer, "%s\r\n", p)
	return n + n1, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.WriterState != WriterStateBody {
		return 0, fmt.Errorf("error: unexpected writerState %d when calling WriteChunkedBodyDone", w.WriterState)
	}
	n, err := w.Writer.Write([]byte("0\r\n"))
	n1, err := w.Writer.Write([]byte("\r\n"))
	return n + n1, err
}
