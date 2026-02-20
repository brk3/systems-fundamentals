package torrentfile

import (
	"io"
	"net/url"
	"strconv"

	be "github.com/jackpal/bencode-go"
	"go-bt-learning.brk3.github.io/internal/bencode"
)

// domain model - decouple ourselves from bencode format specifics
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func NewTorrentFile(r io.Reader) (TorrentFile, error) {
	b := bencode.BencodeTorrent{}
	err := be.Unmarshal(r, &b)
	if err != nil {
		return TorrentFile{}, err
	}
	numPieces := len(b.Info.Pieces) / 20
	pieceHashes := make([][20]byte, numPieces)
	for i := 0; i < numPieces; i++ {
		start := i * 20
		copy(pieceHashes[i][:], b.Info.Pieces[start:start+20])
	}
	tf := TorrentFile{}
	tf.PieceHashes = pieceHashes
	h, err := b.Info.InfoHash()
	if err != nil {
		return TorrentFile{}, err
	}
	tf.InfoHash = h
	tf.Announce = b.Announce
	tf.PieceLength = b.Info.PieceLength
	tf.Length = b.Info.Length
	tf.Name = b.Info.Name
	return tf, nil
}

// buildTrackerURL combines the torrentfile's announce url with several key parameters namely our
// info_hash and peer_id
func (t *TorrentFile) BuildTrackerURL(peerID string, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	infoHash := t.InfoHash
	params := url.Values{
		"info_hash": []string{string(infoHash[:])}, // the file we’re trying to download
		// TODO: find out how to properly pass [20]byte here instead of string
		"peer_id":    []string{peerID}, // 20 byte name to identify ourselves to trackers and peers
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}
