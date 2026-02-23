package torrentfile

import (
	"crypto/sha1"
	"testing"
)

func TestInfoHash(t *testing.T) {
	p1 := sha1.Sum([]byte("piece1"))
	p2 := sha1.Sum([]byte("piece2"))
	pieces := string(p1[:]) + string(p2[:])
	b := bencodeInfo{
		Length:      10,
		Name:        "foo",
		PieceLength: 5,
		Pieces:      pieces,
	}
	want := [20]uint8{237, 244, 120, 216, 145, 14, 59, 209, 182, 21, 51, 112, 162, 150, 171, 205, 224, 159, 244, 129}
	have, err := b.InfoHash()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if want != have {
		t.Errorf("unexpected infohash, have: %x, want: %x", have, want)
	}
}
