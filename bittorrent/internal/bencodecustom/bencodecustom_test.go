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

func TestParseDicct(t *testing.T) {
	r := strings.NewReader("d3:cow3:moo4:spam4:eggse")
	r.ReadByte() // drop leading 'd' to avoid Parse doubling up on itself
	d, err := parseDict(bufio.NewReader(r))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if len(d) != 2 {
		t.Errorf("expected %d items in result list, got %d", 2, len(d))
	}
	val, exists := d["cow"]
	if !exists {
		t.Errorf("missing expected key 'cow' in map")
	}
	if val != "moo" {
		t.Errorf("expected value of 'moo' for key 'cow', got %s", val)
	}
	val, exists = d["spam"]
	if !exists {
		t.Errorf("missing expected key 'spam' in map")
	}
	if d["spam"] != val {
		t.Errorf("expected value of 'eggs' for key 'spam', got %s", val)
	}
}

func TestParse(t *testing.T) {
	input := "d3:barl1:a1:be3:fooi42ee"
	b := bufio.NewReader(strings.NewReader(input))
	resAny, err := Parse(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dict, ok := resAny.(map[string]any)
	if !ok {
		t.Fatalf("expected result to be map[string]any, got %T", resAny)
	}
	fooAny, ok := dict["foo"]
	if !ok {
		t.Fatalf("missing key 'foo'")
	}
	foo, ok := fooAny.(int)
	if !ok || foo != 42 {
		t.Errorf("expected foo to be int 42, got %v", fooAny)
	}
	barAny, ok := dict["bar"]
	if !ok {
		t.Fatalf("missing key 'bar'")
	}
	barList, ok := barAny.([]any)
	if !ok || len(barList) != 2 {
		t.Fatalf("expected bar to be a []any of length 2, got %v", barAny)
	}
	if a, ok := barList[0].(string); !ok || a != "a" {
		t.Errorf("expected bar[0] to be 'a', got %v", barList[0])
	}
	if bVal, ok := barList[1].(string); !ok || bVal != "b" {
		t.Errorf("expected bar[1] to be 'b', got %v", barList[1])
	}
}
