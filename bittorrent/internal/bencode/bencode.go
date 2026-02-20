package bencode

import (
	"bytes"
	"crypto/sha1"

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

type BencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int         `bencode:"creation date"`
	Info         bencodeInfo `bencode:"info"`
}

// infoHash uniquely identifies files when we talk to trackers and peers
func (i bencodeInfo) InfoHash() ([hashLen]byte, error) {
	buf := bytes.Buffer{}
	err := bencode.Marshal(&buf, i)
	if err != nil {
		return [hashLen]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), err
}
