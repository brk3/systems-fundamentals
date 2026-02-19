package bittorrent

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Handshake is a special message that a peer uses to identify itself
type Handshake struct {
	Pstr     string // protocol identifier
	InfoHash [20]byte
	PeerID   [20]byte
}

// Unmarshal parses peer IP addresses and ports from a buffer
func Unmarshal(peersBin []byte) ([]Peer, error) {
	const ipSize = 4
	const portSize = 2
	const peerSize = ipSize + portSize
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := fmt.Errorf("received malformed peers, len %d", len(peersBin))
		return nil, err
	}
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+ipSize])
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+ipSize : offset+peerSize])
	}
	return peers, nil
}

func (p *Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

func (p *Peer) Connect() (io.ReadWriteCloser, error) {
	conn, err := net.DialTimeout("tcp", p.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (p *Peer) Handshake(rw io.ReadWriter, h Handshake) (Handshake, error) {
	_, err := rw.Write(h.Serialize())
	if err != nil {
		return Handshake{}, err
	}
	res := make([]byte, 68)
	_, err = io.ReadFull(rw, res)
	if err != nil {
		return Handshake{}, err
	}
	hr := Handshake{}
	hr.Deserialize(res)
	return hr, nil
}

// Serialize serializes the handshake to a buffer
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

// Read parses a handshake from a stream
func (h *Handshake) Deserialize(buf []byte) {
	pstrLen := int(buf[0])
	buf = buf[1:]
	h.Pstr = string(buf[:pstrLen])
	buf = buf[pstrLen:]
	buf = buf[8:] // skip reserved
	copy(h.InfoHash[:], buf[:20])
	buf = buf[20:]
	copy(h.PeerID[:], buf[:20])
}

func NewHandshake(infoHash [20]byte) Handshake {
	peerID := [20]byte{}
	copy(peerID[:], PeerID)
	return Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}
