package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(r string) (RequestLine, error) {
	parts := strings.Split(r, " ")
	if len(parts) != 3 {
		return RequestLine{}, fmt.Errorf("invalid request line '%s'", r)
	}

	method := parts[0]
	for _, c := range method {
		if !unicode.IsUpper(c) || !unicode.IsLetter(c) {
			return RequestLine{}, fmt.Errorf("method: '%s' is not valid", method)
		}
	}

	version := strings.Split(parts[len(parts)-1], "/")[1]
	if version != "1.1" {
		return RequestLine{}, fmt.Errorf("version: '%s' is not valid", version)
	}

	return RequestLine{
		HttpVersion:   version,
		RequestTarget: parts[1],
		Method:        method,
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	b, err := io.ReadAll(reader)
	lines := strings.Split(string(b[:]), "\r\n")

	rl, err := parseRequestLine(lines[0])
	if err != nil {
		return &Request{rl}, err
	}

	return &Request{rl}, nil
}
