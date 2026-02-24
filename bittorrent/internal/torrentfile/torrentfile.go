package torrentfile

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"go-bt-learning.brk3.github.io/internal/bencodecustom"
)

// serialisation structs - directly maps to torrentfile spec
type bencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int         `bencode:"creation date"`
	Info         bencodeInfo `bencode:"info"`
}

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
	b, err := unmarshal(bufio.NewReader(r))
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
	tf.InfoHash = sha1.Sum(b.Info.marshal())
	tf.Announce = b.Announce
	tf.PieceLength = b.Info.PieceLength
	tf.Length = b.Info.Length
	tf.Name = b.Info.Name
	return tf, nil
}

func unmarshal(b *bufio.Reader) (bencodeTorrent, error) {
	val, err := bencodecustom.Parse(b)
	if err != nil {
		return bencodeTorrent{}, err
	}
	t, ok := val.(map[string]any)
	if !ok {
		return bencodeTorrent{}, fmt.Errorf("error converting Parse response to map[string]any")
	}
	info, ok := t["info"].(map[string]any)
	if !ok {
		return bencodeTorrent{}, fmt.Errorf("error converting Info from response to map[string]any")
	}
	bt := bencodeTorrent{
		Announce: t["announce"].(string),
		Info: bencodeInfo{
			Length:      info["length"].(int),
			Name:        info["name"].(string),
			PieceLength: info["piece length"].(int),
			Pieces:      info["pieces"].(string),
		},
	}
	return bt, nil
}

func (i bencodeInfo) marshal() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('d')
	buf.WriteString(fmt.Sprintf("6:lengthi%se", strconv.Itoa(i.Length)))
	buf.WriteString(fmt.Sprintf("4:name%s:%s", strconv.Itoa(len(i.Name)), i.Name))
	buf.WriteString(fmt.Sprintf("12:piece lengthi%se", strconv.Itoa(i.PieceLength)))
	buf.WriteString(fmt.Sprintf("6:pieces%s:%s", strconv.Itoa(len(i.Pieces)), i.Pieces))
	buf.WriteByte('e')
	return buf.Bytes()
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
