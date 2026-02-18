package bittorrent

import (
	"bufio"
	"os"
	"testing"

	bencode "github.com/jackpal/bencode-go"
)

func TestToTorrentFile(t *testing.T) {
	// TODO
	f, err := os.Open("../../debian-11.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		t.Errorf("error opening torrent file: %v", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	b := bencodeTorrent{}
	err = bencode.Unmarshal(r, &b)
	if err != nil {
		t.Errorf("error unmarshalling torrent file: %v", err)
	}

	// TODO
	// tf, err := b.toTorrentFile()
	// if err != nil {
	// 	t.Errorf("error converting bencodeTorrent to TorrentFile file: %v", err)
	// }
	// want := 1528
	// have := len(tf.PieceHashes)
	// if want != have {
	// 	t.Errorf("expected %d piece hashes, got %d", want, have)
	// }
}
