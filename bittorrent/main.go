package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

// Currently ignoring the 'TorrentFile' abstraction described in article

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

func (i bencodeInfo) infoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, i)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), err
}

func (t bencodeTorrent) trackerURL(peerID string, port int) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	infoHash, err := t.Info.infoHash()
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{peerID},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Info.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func main() {
	f, err := os.Open("debian-11.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Println("error opening torrent file: ", err)
		os.Exit(1)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	torrent := bencodeTorrent{}
	err = bencode.Unmarshal(r, &torrent)
	if err != nil {
		fmt.Println("error unmarshalling torrent file: ", err)
		os.Exit(1)
	}

	infoHash, err := torrent.Info.infoHash()
	if err != nil {
		fmt.Println("error generating infoHash: ", err)
		os.Exit(1)
	}
	fmt.Printf("infoHash: %x\n", infoHash)

	trackerURL, err := torrent.trackerURL("paulstorrentclient!", 8080)
	if err != nil {
		fmt.Println("error generating trackerURL: ", err)
		os.Exit(1)
	}
	fmt.Println("trackerURL: ", trackerURL)
}
