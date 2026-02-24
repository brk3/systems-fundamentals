package torrent

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-bt-learning.brk3.github.io/internal/bencodecustom"
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
	tr, err := unmarshalTrackerResponse(res.Body)
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

func unmarshalTrackerResponse(r io.Reader) (trackerResponse, error) {
	val, err := bencodecustom.Parse(bufio.NewReader(r))
	if err != nil {
		return trackerResponse{}, err
	}
	rawMap, ok := val.(map[string]any)
	if !ok {
		return trackerResponse{}, fmt.Errorf("error converting tracker response to map[string]any")
	}
	if reason, err := safeGet[string](rawMap, "failure reason"); err == nil {
		return trackerResponse{FailureReason: reason}, nil
	}
	t := trackerResponse{}
	interval, err := safeGet[int](rawMap, "interval")
	if err != nil {
		return t, err
	}
	t.Interval = interval
	peers, err := safeGet[string](rawMap, "peers")
	if err != nil {
		return t, err
	}
	t.Peers = peers
	return t, nil
}

func safeGet[T any](m map[string]any, key string) (T, error) {
	var zero T
	val, found := m[key]
	if !found {
		return zero, fmt.Errorf("key '%s' not found in map", key)
	}
	typedVal, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("key '%s' exists but is type %T, not the expected type", key, val)
	}
	return typedVal, nil
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
		fmt.Printf("%s: error creating client for peer: %v\n", peer.String(), err)
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
			err = checkIntegrity(pw, buf)
			if err != nil {
				fmt.Printf("%s: piece #%d failed integrity check, requeueing\n", peer.String(), pw.index)
				workQueue <- pw
				continue
			}
			resQueue <- pieceResult{index: pw.index, buf: buf}
		} else {
			_, err := c.HandleMessage()
			if err != nil {
				fmt.Printf("%s: error reading message from peer: %v\n", peer.String(), err)
				return
			}
		}
	}
}

func checkIntegrity(pw pieceWork, buf []byte) error {
	if s := sha1.Sum(buf); s != pw.hash {
		return fmt.Errorf("received piece hash (%s) doesn't match expected (%s)\n", s, pw.hash)
	}
	return nil
}

func downloadPiece(c *client.Client, pw pieceWork) ([]byte, error) {
	piece := pieceProgress{
		client: c,
		buf:    make([]byte, pw.length),
	}
	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline
	for piece.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !piece.client.Choked {
			for piece.backlog < MaxBacklog && piece.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-piece.requested < blockSize {
					blockSize = pw.length - piece.requested
				}
				err := c.SendRequest(pw.index, piece.requested, blockSize)
				if err != nil {
					return nil, err
				}
				piece.backlog++
				piece.requested += blockSize
			}
		}
		msg, err := piece.client.HandleMessage()
		if err != nil {
			return nil, err
		}
		if msg != nil && msg.ID == message.MsgPiece {
			n := message.ParsePiece(piece.buf, msg)
			piece.downloaded += n
			piece.backlog--
		}
	}
	fmt.Printf("%s: successfully downloaded piece %d, size %d\n", c.Peer.String(), pw.index, len(piece.buf))
	return piece.buf, nil
}

func (t *Torrent) Download() []byte {
	workQueue := make(chan pieceWork, len(t.File.PieceHashes)) // buffered channel
	resQueue := make(chan pieceResult)
	fmt.Printf("we have %d pieces to fetch\n", len(t.File.PieceHashes))
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
	return buf
}
