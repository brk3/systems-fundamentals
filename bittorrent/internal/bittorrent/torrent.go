package bittorrent

import (
	"fmt"
	"io"
	"net/http"
	"time"

	bencode "github.com/jackpal/bencode-go"
)

const PeerID = "paulsbittorentclient" // 20 chars

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

func (t *Torrent) waitForBitfield(r io.Reader) error {
	for {
		m, err := ReadMessage(r)
		if err != nil {
			return err
		}
		if m.ID == MsgBitfield {
			t.Bitfield = m.Serialize()
			return nil
		}
		fmt.Printf("received message ID %d while waiting for bitfield", m.ID)
	}
}

func (t *Torrent) startDownloadWorker(peer Peer, workQueue chan pieceWork, resQueue chan pieceResult) {
	conn, err := peer.Connect()
	if err != nil {
		fmt.Printf("%s: error connecting to peer: %v\n", peer.String(), err)
		return
	}
	fmt.Printf("%s: connected\n", peer.String())
	defer conn.Close()
	h := NewHandshake(t.File.InfoHash)
	rh, err := peer.Handshake(conn, h)
	if err != nil {
		fmt.Printf("%s: error handshaking with peer: %v\n", peer.String(), err)
		return
	}
	if rh.InfoHash != t.File.InfoHash {
		fmt.Printf("%s: infohash from peer (%s) doesn't match what we asked for (%s)\n",
			peer.String(), rh.InfoHash, t.File.InfoHash)
		return
	}
	err = t.waitForBitfield(conn)
	if err != nil {
		fmt.Printf("%s: error reading bitfield from peer: %v\n", peer.String(), err)
		return
	}
	// sendinterested
	for pw := range workQueue {
		resQueue <- pieceResult{index: pw.index} // simulate results
	}
}

func (t *Torrent) Download() {
	workQueue := make(chan pieceWork, len(t.File.PieceHashes)) // buffered channel
	resQueue := make(chan pieceResult)
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
		if res.err != nil {
			// TODO: handle worker failure
			fmt.Printf("worker error downloading piece %d: %v\n", res.index, res.err)
			continue
		}
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
	}
}
