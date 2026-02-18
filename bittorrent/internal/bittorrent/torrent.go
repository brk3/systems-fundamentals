package bittorrent

import (
	"fmt"
	"net/http"
	"time"

	bencode "github.com/jackpal/bencode-go"
)

type Torrent struct {
	File     TorrentFile
	Peers    []Peer
	Bitfield Bitfield
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
	err   error
}

func NewTorrent(t TorrentFile) *Torrent {
	return &Torrent{
		File:     t,
		Bitfield: make(Bitfield, (len(t.PieceHashes)+7)/8), // round up trick to ensure enough bytes
	}
}

type trackerResponse struct {
	FailureReason string `bencode:"failure reason"`
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"` // contains binary data (compact format)
}

func (t *Torrent) Announce(peerID string, port uint16) error {
	tu, err := t.File.BuildTrackerURL(peerID, port)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", tu, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("tracker returned non-200 status: %d", res.StatusCode)
	}
	var tr trackerResponse
	err = bencode.Unmarshal(res.Body, &tr)
	if err != nil {
		return fmt.Errorf("error decoding tracker response: %w", err)
	}
	if tr.FailureReason != "" {
		return fmt.Errorf("tracker failed: %s", tr.FailureReason)
	}
	peers, err := Unmarshal([]byte(tr.Peers))
	if err != nil {
		return fmt.Errorf("error parsing peers: %w", err)
	}
	t.Peers = peers
	return nil
}

func (t *Torrent) calculatePieceSize(index int) int {
	remainder := t.File.Length % t.File.PieceLength
	if remainder > 0 && index == len(t.File.PieceHashes)-1 {
		return remainder
	}
	return t.File.PieceLength
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin, end int) {
	prevEnd := t.File.PieceLength * index
	return prevEnd, prevEnd + t.calculatePieceSize(index)
}

// TODO
func (t *Torrent) startDownloadWorker(peer Peer, workQueue chan pieceWork, resQueue chan pieceResult) {
	for pw := range workQueue { // pw := <-workQueue
		resQueue <- pieceResult{index: pw.index}
	}
	// open tcp conn with peer
	// do handshake, receive bitfield
	// take a piece of work from queue
	// does peer have this piece
	// if no, put back on queue
	// if yes, try download
	// download ok? check hash
	// hash ok? send result to channel
	// ...
}

func (t *Torrent) Download() {
	workQueue := make(chan pieceWork, len(t.File.PieceHashes)) // buffered channel
	resQueue := make(chan pieceResult)
	// TODO: should we check bitfield here?
	for index, hash := range t.File.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- pieceWork{index, hash, length}
	}
	close(workQueue)
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, resQueue)
	}
	buf := make([]byte, t.File.Length)
	donePieces := 0
	for donePieces < len(t.File.PieceHashes) {
		res := <-resQueue
		// TODO: check for error on pieceResult
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
	}
}
