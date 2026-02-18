package bittorrent

import (
	"testing"
	"net"
)

func TestCalculatePieceSize(t *testing.T) {
	// last piece of even size
	tf := TorrentFile{
		PieceHashes: make([][20]byte, 4),
		PieceLength: 25,
		Length: 100,
	}
	to := NewTorrent(tf)
	want := 25
	have := to.calculatePieceSize(3)
	if have != want {
		t.Errorf("expected piece size of '%d', got '%d'", want, have)
	}

	// last piece of odd size
	tf = TorrentFile{
		PieceHashes: make([][20]byte, 4),
		PieceLength: 25,
		Length: 110,
	}
	to = NewTorrent(tf)
	want = 10
	have = to.calculatePieceSize(3)
	if have != want {
		t.Errorf("expected piece size of '%d', got '%d'", want, have)
	}

	// standard piece
	tf = TorrentFile{
		PieceHashes: make([][20]byte, 4),
		PieceLength: 25,
		Length: 110,
	}
	to = NewTorrent(tf)
	want = 25
	have = to.calculatePieceSize(1)
	if have != want {
		t.Errorf("expected piece size of '%d', got '%d'", want, have)
	}
}

func TestCalculateBoundsForPiece(t *testing.T) {
	// first piece
	tf := TorrentFile{
		PieceHashes: make([][20]byte, 4),
		PieceLength: 256,
		Length: 1000,
	}
	to := NewTorrent(tf)
	want_start, want_end := 0, 256
	have_start, have_end := to.calculateBoundsForPiece(0)
	if have_start != want_start || have_end != want_end {
		t.Errorf("expected (start, end) of (%d, %d), got (%d, %d)",
		want_start, want_end, have_start, have_end)
	}

	// middle piece
	want_start, want_end = 256, 512
	have_start, have_end = to.calculateBoundsForPiece(1)
	if have_start != want_start || have_end != want_end {
		t.Errorf("expected (start, end) of (%d, %d), got (%d, %d)",
		want_start, want_end, have_start, have_end)
	}

	// last piece
	want_start, want_end = 768, 1000
	have_start, have_end = to.calculateBoundsForPiece(3)
	if have_start != want_start || have_end != want_end {
		t.Errorf("expected (start, end) of (%d, %d), got (%d, %d)",
		want_start, want_end, have_start, have_end)
	}
}

func TestDownload(t *testing.T) {
	tf := TorrentFile{
		PieceHashes: make([][20]byte, 5),
		PieceLength: 2,
		Length: 10,
	}
	to := NewTorrent(tf)
	to.Peers = []Peer{ { IP: net.ParseIP("1.2.3.4"), Port: 6881, } }
	to.Download()
}