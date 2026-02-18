package bittorrent

import (
	"io"
	"net/url"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

// domain model - decouple ourselves from bencode format specifics
type TorrentFile struct {
	Announce    string
	InfoHash    [hashLen]byte
	PieceHashes [][hashLen]byte
	PieceLength int
	Length      int
	Name        string
}

func NewTorrentFile(r io.Reader) (TorrentFile, error) {
	b := bencodeTorrent{}
	err := bencode.Unmarshal(r, &b)
	if err != nil {
		return TorrentFile{}, err
	}
	numPieces := len(b.Info.Pieces) / hashLen
	pieceHashes := make([][hashLen]byte, numPieces)
	for i := 0; i < numPieces; i++ {
		start := i * hashLen
		copy(pieceHashes[i][:], b.Info.Pieces[start:start+hashLen])
	}
	tf := TorrentFile{}
	tf.PieceHashes = pieceHashes
	h, err := b.Info.infoHash()
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
