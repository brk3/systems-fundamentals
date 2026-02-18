package bittorrent

// domain model - decouple ourselves from bencode format specifics
type TorrentFile struct {
	Announce    string
	InfoHash    [hashLen]byte
	PieceHashes [][hashLen]byte
	PieceLength int
	Length      int
	Name        string
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

type Torrent struct {
	File     TorrentFile
	Peers    []Peer
	Bitfield Bitfield
}

func NewTorrent(t TorrentFile) *Torrent {
	return &Torrent{
		File:     t,
		Bitfield: make(Bitfield, (len(t.PieceHashes)+7)/8), // round up trick to ensure enough bytes
	}
}

func (t *Torrent) calculatePieceSize(index int) int {
	remainder := t.File.Length % t.File.PieceLength
	if remainder > 0 && index == len(t.File.PieceHashes)-1 {
		return remainder
	}
	return t.File.PieceLength
}

// TODO
func (t *Torrent) calculateBoundsForPiece(index int) (begin, end int) {
	return 0, 0
}

// TODO
func (t *Torrent) startDownloadWorker(peer Peer, workQueue chan pieceWork, resQueue chan pieceResult) {
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
