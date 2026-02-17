package bittorrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/url"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

const hashLen = 20 // sha1

// serialisation structs - directly maps to bencode spec
type bencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int         `bencode:"creation date"`
	Info         bencodeInfo `bencode:"info"`
}

func (b bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	if len(b.Info.Pieces)%hashLen != 0 {
		return TorrentFile{}, fmt.Errorf("received malformed pieces, len %d", len(b.Info.Pieces))
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

// infoHash uniquely identifies files when we talk to trackers and peers
func (i bencodeInfo) infoHash() ([20]byte, error) {
	buf := bytes.Buffer{}
	err := bencode.Marshal(&buf, i)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), err
}

func (t bencodeTorrent) trackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	infoHash, err := t.Info.infoHash()
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(infoHash[:])}, // the file we’re trying to download
		"peer_id":    []string{string(peerID[:])},   // 20 byte name to identify ourselves to trackers and peers
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Info.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}
