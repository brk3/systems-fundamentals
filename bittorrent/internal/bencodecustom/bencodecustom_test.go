package bencodecustom

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestParseInt(t *testing.T) {
	want := 52
	s := strings.NewReader(fmt.Sprintf("i%de", want))
	have, err := parseInt(bufio.NewReader(s))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if have != want {
		t.Errorf("expected %d, got %d", want, have)
	}
}

func TestParseString(t *testing.T) {
	// standard
	want := "hello!"
	s := strings.NewReader(fmt.Sprintf("%d:%s", len(want), want))
	have, err := parseString(bufio.NewReader(s))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if have != want {
		t.Errorf("expected %s, got %s", want, have)
	}

	// containing colon
	want = "hel:lo"
	s = strings.NewReader(fmt.Sprintf("%d:%s", len(want), want))
	have, err = parseString(bufio.NewReader(s))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if have != want {
		t.Errorf("expected %s, got %s", want, have)
	}

	// empty
	want = ""
	s = strings.NewReader(fmt.Sprintf("%d:%s", len(want), want))
	have, err = parseString(bufio.NewReader(s))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if have != want {
		t.Errorf("expected %s, got %s", want, have)
	}
}

func TestParseList(t *testing.T) {
	// two strings
	r := strings.NewReader("l4:spam4:eggse")
	r.ReadByte() // drop leading 'l' to avoid Parse doubling up on itself
	l, err := parseList(bufio.NewReader(r))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if len(l) != 2 {
		t.Errorf("expected %d items in result list, got %d", 2, len(l))
	}
	if val, ok := l[0].(string); !ok || val != "spam" {
		t.Errorf("expected first item to be 'spam', got %v", l[0])
	}
	if val, ok := l[1].(string); !ok || val != "eggs" {
		t.Errorf("expected second item to be 'eggs', got %v", l[1])
	}

	// mix of string and int
	r = strings.NewReader("l4:spami1e4:eggse")
	r.ReadByte() // drop leading 'l' to avoid Parse doubling up on itself
	l, err = parseList(bufio.NewReader(r))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if len(l) != 3 {
		t.Errorf("expected %d items in result list, got %d", 3, len(l))
	}
	if val, ok := l[0].(string); !ok || val != "spam" {
		t.Errorf("expected first item to be 'spam', got %v", l[0])
	}
	if val, ok := l[1].(int); !ok || val != 1 {
		t.Errorf("expected second item to be 1 got %v", l[1])
	}
	if val, ok := l[2].(string); !ok || val != "eggs" {
		t.Errorf("expected second item to be 'eggs', got %v", l[2])
	}
}
