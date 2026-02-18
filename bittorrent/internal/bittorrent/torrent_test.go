package bittorrent

import "testing"

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
