package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type ParseState int

const (
	initialised ParseState = iota
	done
)

const bufferSize int = 8

type Request struct {
	RequestLine RequestLine
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

func (r *Request) parse(data []byte) (int, error) {
	if r.ParseState == initialised {
		rl, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = rl
		r.ParseState = done

		return n, nil
	} else if r.ParseState == done {
		return 0, fmt.Errorf("trying to read data in a done state")
	}

	return 0, fmt.Errorf("unknown state")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	requestLineLength := 0
	r := Request{ParseState: initialised}

	for r.ParseState != done {
		if readToIndex == len(buf) {
			tmp := make([]byte, len(buf)*2)
			copy(tmp, buf)
			buf = tmp
		}

		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			r.ParseState = done
			break
		}
		readToIndex += n

		requestLineLength, err = r.parse(buf[:readToIndex])
		if err != nil {
			return &r, err
		}
	}

	tmp := make([]byte, readToIndex-requestLineLength)
	copy(tmp, buf[requestLineLength:])
	buf = tmp
	readToIndex -= requestLineLength

	return &r, nil
}
