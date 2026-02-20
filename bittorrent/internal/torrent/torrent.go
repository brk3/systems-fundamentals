package torrent

import (
	"fmt"
	"net/http"
	"time"

	bencode "github.com/jackpal/bencode-go"
	"go-bt-learning.brk3.github.io/internal/bitfield"
	"go-bt-learning.brk3.github.io/internal/client"
	"go-bt-learning.brk3.github.io/internal/message"
	"go-bt-learning.brk3.github.io/internal/peer"
	"go-bt-learning.brk3.github.io/internal/torrentfile"
)

const (
	// MaxBlockSize is the largest number of bytes a request can ask for
	MaxBlockSize = 16384

	// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
	MaxBacklog = 5
)

type Torrent struct {
	File     torrentfile.TorrentFile
	Peers    []peer.Peer
	Bitfield bitfield.Bitfield
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

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

type trackerResponse struct {
	FailureReason string `bencode:"failure reason"`
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"`
}

func NewTorrent(t torrentfile.TorrentFile) *Torrent {
	return &Torrent{
		File:     t,
		Bitfield: make(bitfield.Bitfield, (len(t.PieceHashes)+7)/8), // round up trick to ensure enough bytes
	}
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
	peers, err := peer.Unmarshal([]byte(tr.Peers))
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

func (t *Torrent) startDownloadWorker(peer peer.Peer, workQueue chan pieceWork, resQueue chan pieceResult) {
	c, err := client.NewClient(peer, t.File.InfoHash)
	if err != nil {
		fmt.Printf("%s: error creating client for peer: %v", peer.String(), err)
		return
	}
	defer c.Conn.Close()
	c.Conn.Write((&message.Message{ID: message.MsgInterested}).Serialize())
	for {
		if !c.Choked && c.Bitfield != nil {
			pw, ok := <-workQueue
			if !ok {
				fmt.Printf("%s: no more work in queue, closing\n", peer.String())
				return
			}
			if !c.Bitfield.HasPiece(pw.index) {
				fmt.Printf("%s: peer doesn't have piece index %d, requeuing\n", peer.String(), pw.index)
				workQueue <- pw
				continue
			}
			buf, err := downloadPiece(c, pw)
			if err != nil {
				fmt.Printf("%s: error downloading piece index %d, requeuing\n", peer.String(), pw.index)
				workQueue <- pw
				continue
			}
			resQueue <- pieceResult{index: pw.index, buf: buf}
		}
		err := c.HandleMessage()
		if err != nil {
			fmt.Printf("%s: error reading message from peer: %v\n", peer.String(), err)
			return
		}
	}
}

func downloadPiece(c *client.Client, pw pieceWork) ([]byte, error) {
	state := pieceProgress{
		client: c,
		buf:    make([]byte, pw.length),
	}

	// TODO
	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	// c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	// defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.client.HandleMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (t *Torrent) Download() {
	workQueue := make(chan pieceWork, len(t.File.PieceHashes)) // buffered channel
	resQueue := make(chan pieceResult)
	for index, hash := range t.File.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- pieceWork{index, hash, length}
	}
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
	close(workQueue)
}
