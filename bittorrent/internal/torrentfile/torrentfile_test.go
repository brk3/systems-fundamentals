package torrentfile

import (
	"crypto/sha1"
	"testing"
)

func TestMarshal(t *testing.T) {
	p1 := sha1.Sum([]byte("piece1"))
	p2 := sha1.Sum([]byte("piece2"))
	b := bencodeInfo{
		Length:      10,
		Name:        "foo",
		PieceLength: 5,
		Pieces:      string(p1[:]) + string(p2[:]),
	}
	want := [20]uint8{237, 244, 120, 216, 145, 14, 59, 209, 182, 21, 51, 112, 162, 150, 171, 205, 224, 159, 244, 129}
	have := sha1.Sum(b.marshal())
	if want != have {
		t.Errorf("unexpected infohash, have: %x, want: %x", have, want)
	}
}
