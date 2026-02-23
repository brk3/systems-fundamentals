package bencodecustom

import (
	"bufio"
	"fmt"
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
		i, err := parseInt(b)
		return i, err
	case 'l':
		l, err := parseList(b)
		return l, err
	case 'd':
		d, err := parseDict(b)
		return d, err
	default:
		return nil, fmt.Errorf("unknown type found in parse: %v", t[0])
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

func parseDict(b *bufio.Reader) (map[string]any, error) {
	t, err := b.Peek(1)
	if err != nil {
		return nil, err
	}
	res := map[string]any{}
	for t[0] != 'e' {
		key, err := parseString(b)
		if err != nil {
			return nil, err
		}
		val, err := Parse(b)
		if err != nil {
			return nil, err
		}
		_, exists := res[key]
		if exists {
			return nil, fmt.Errorf("dupe key '%s' found in dict", key)
		}
		res[key] = val
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
