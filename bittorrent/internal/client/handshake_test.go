package client

import (
	"testing"
)

func TestHandshakeRoundTrip(t *testing.T) {
	h1 := Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}
	// test round trip
	data := h1.Serialize()
	h2 := Handshake{}
	h2.Deserialize(data)
	if h1.Pstr != h2.Pstr {
		t.Errorf("Pstr mismatch: got %q, want %q", h2.Pstr, h1.Pstr)
	}
	if h1.InfoHash != h2.InfoHash {
		t.Errorf("InfoHash mismatch: got %x, want %x", h2.InfoHash, h1.InfoHash)
	}
	if h1.PeerID != h2.PeerID {
		t.Errorf("PeerID mismatch: got %x, want %x", h2.PeerID, h1.PeerID)
	}
}
