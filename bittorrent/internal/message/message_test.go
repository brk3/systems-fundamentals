package message

import (
	"bytes"
	"testing"
)

func TestSerializeKeepalive(t *testing.T) {
	var m *Message
	want := []byte{0,0,0,0}
	have := m.Serialize()
	if !bytes.Equal(have, want) {
		t.Errorf("expected %v when serialised keepalive, got %v", want, have)
	}
}

func TestSerializeNoPayload(t *testing.T) {
	m1 := Message{ ID: MsgChoke }
	r := bytes.NewReader(m1.Serialize())
	m2, err := ReadMessage(r)
	if err != nil {
		t.Errorf("unexpected error in ReadMessage: %v", err)
	}
	if m2.ID != MsgChoke {
		t.Errorf("expected ID %d, got %d", m1.ID, m2.ID)
	}
	if len(m2.Payload) != 0 {
		t.Errorf("expected no payload, got %v", m2.Payload)
	}
}