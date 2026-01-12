package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	clrfIndex := bytes.Index(data, []byte("\r\n"))
	if clrfIndex == -1 {
		// need more data
		return 0, false, nil
	}
	if clrfIndex == 0 {
		// hit the final CRLF, done
		return 2, true, nil
	}

	rawHeader := data[:clrfIndex]
	key, value, found := bytes.Cut(rawHeader, []byte(":"))
	if !found {
		return 0, false, fmt.Errorf("invalid header '%s'", rawHeader)
	}

	if !validFieldName(key) {
		return 0, false, fmt.Errorf("invalid header '%s'", rawHeader)
	}

	// check for white space before key or after colon
	if key[0] == ' ' || key[len(key)-1] == ' ' {
		return 0, false, fmt.Errorf("invalid header '%s'", rawHeader)
	}

	k := strings.ToLower(strings.TrimSpace(string(key)))
	val, found := h[k]
	if found {
		h[k] = strings.Join([]string{val, strings.TrimSpace(string(value))}, ", ")
	} else {
		h[k] = strings.TrimSpace(string(value))
	}
	return clrfIndex + 2, false, nil
}

func validFieldName(name []byte) bool {
	if len(name) < 1 {
		return false
	}
	for _, b := range name {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
			continue
		} else {
			switch b {
			case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
				continue
			}
		}
		return false
	}
	return true
}
