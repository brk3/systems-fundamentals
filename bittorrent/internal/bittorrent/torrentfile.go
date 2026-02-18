package bittorrent

import (
	"io"

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