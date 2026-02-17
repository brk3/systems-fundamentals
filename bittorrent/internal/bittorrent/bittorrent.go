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

type Torrent struct {
}

func (t *Torrent) Download() ([]byte, error) {
	// TODO
	return nil, nil
}
