package bencodecustom

import (
	"bufio"
	"io"
	"strconv"
)

func Parse(b *bufio.Reader) (any, error) {
	t, err := b.Peek(1)
	if err != nil {
		return nil, err
	}
	if t[0] >= '0' && t[0] <= '9' {
		s, err := parseString(b)
		return s, err
	}
	_, err = b.ReadByte()
	if err != nil {
		return nil, err
	}
	switch t[0] {
	case 'i':
		s, err := parseInt(b)
		return s, err
	case 'l':
		s, err := parseList(b)
		return s, err
	default:
		return "not implemented", nil
	}
}

func parseList(b *bufio.Reader) ([]any, error) {
	t, err := b.Peek(1)
	if err != nil {
		return nil, err
	}
	res := []any{}
	for t[0] != 'e' {
		item, err := Parse(b)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
		t, err = b.Peek(1)
		if err != nil {
			return nil, err
		}
	}
	b.ReadByte() // discard final 'e'
	return res, nil
}

func parseString(b *bufio.Reader) (string, error) {
	l, err := b.ReadString(':')
	if err != nil {
		return "", err
	}
	sLen, err := strconv.Atoi(l[:len(l)-1])
	buf := make([]byte, sLen)
	_, err = io.ReadFull(b, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func parseInt(b *bufio.Reader) (int, error) {
	s, err := b.ReadString('e')
	if err != nil {
		return 0, err
	}
	val := s[:len(s)-1]
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return n, nil
}
