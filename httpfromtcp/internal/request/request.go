package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"httpfromtcp/internal/headers"
)

type ParseState int

const (
	initialised ParseState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

const bufferSize int = 8

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	ParseState  ParseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(r string) (RequestLine, int, error) {
	lines := strings.Split(r, "\r\n")
	if len(lines) == 1 {
		return RequestLine{}, 0, nil
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line '%s'", r)
	}

	method := parts[0]
	for _, c := range method {
		if !unicode.IsUpper(c) || !unicode.IsLetter(c) {
			return RequestLine{}, 0, fmt.Errorf("method: '%s' is not valid", method)
		}
	}

	version := strings.Split(parts[len(parts)-1], "/")[1]
	if version != "1.1" {
		return RequestLine{}, 0, fmt.Errorf("version: '%s' is not valid", version)
	}

	requestTarget := parts[1]

	return RequestLine{
		HttpVersion:   version,
		RequestTarget: requestTarget,
		Method:        method,
	}, len(lines[0]) + 2, nil // + 2 to include \r\n
}

func (r *Request) parseBody(data []byte, contentLength int) ([]byte, int, error) {
	if len(data) > contentLength {
		return nil, 0, fmt.Errorf("got more data than specified by content-length header")
	}
	if len(data) == contentLength {
		return data, contentLength, nil
	}
	return nil, 0, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.ParseState {
	case initialised:
		rl, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		} else if n == 0 {
			return 0, nil
		} else {
			r.RequestLine = rl
			r.ParseState = requestStateParsingHeaders
		}
		return n, nil
	case requestStateParsingHeaders:
		for {
			n, done, err := r.Headers.Parse(data)
			if err != nil {
				return 0, err
			}
			if done {
				r.ParseState = requestStateParsingBody
				return 2, nil
			}
			return n, nil
		}
	case requestStateParsingBody:
		cl, found := r.Headers.Get("content-length")
		if !found {
			r.ParseState = requestStateDone
			return 0, nil
		}
		clInt, err := strconv.Atoi(cl)
		if err != nil {
			return 0, err
		}
		body, n, err := r.parseBody(data, clInt)
		if err != nil {
			return 0, err
		} else if n == 0 {
			return 0, nil
		} else {
			r.Body = body
			r.ParseState = requestStateDone
		}
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("trying to read data in a done state")
	}

	return 0, fmt.Errorf("unknown state")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	r := Request{
		ParseState: initialised,
		Headers:    headers.Headers{},
	}

	for r.ParseState != requestStateDone {
		if readToIndex == len(buf) {
			tmp := make([]byte, len(buf)*2)
			copy(tmp, buf)
			buf = tmp
		}

		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n
		if err == io.EOF {
			// final parse if anything left in buffer
			n, err = r.parse(buf[:readToIndex])
			if err != nil {
				return nil, err
			}
			if r.ParseState != requestStateDone {
				return nil, fmt.Errorf("unexpected EOF, request incomplete")
			}
			break
		}

		n, err = r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		if n > 0 {
			copy(buf, buf[n:readToIndex])
			readToIndex -= n
		}
	}

	return &r, nil
}
