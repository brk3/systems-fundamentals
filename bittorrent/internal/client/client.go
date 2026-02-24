package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"go-bt-learning.brk3.github.io/internal/bitfield"
	"go-bt-learning.brk3.github.io/internal/message"
)

// PeerID is a 20 byte identifier for our client
const PeerID = "paulsbittorentclient"

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	Peer     Peer
}

func NewClient(peer Peer, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	p := [20]byte{}
	copy(p[:], PeerID)
	h := Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   p,
	}
	_, err = doHandshake(conn, h, peer)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: nil,
		Peer:     peer,
	}, nil
}

// HandleMessage updates the Client state based on the message received. It returns the message for
// optional further processing.
func (c *Client) HandleMessage() (*message.Message, error) {
	msg, err := message.ReadMessage(c.Conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		fmt.Printf("%s: received keepalive message\n", c.Peer.String())
		return nil, nil
	}
	switch msg.ID {
	case message.MsgBitfield:
		fmt.Printf("%s: received bitfield message\n", c.Peer.String())
		c.Bitfield = msg.Payload
	case message.MsgUnchoke:
		fmt.Printf("%s: received unchoke message\n", c.Peer.String())
		c.Choked = false
	case message.MsgChoke:
		fmt.Printf("%s: received choke message\n", c.Peer.String())
		c.Choked = true
	case message.MsgHave:
		fmt.Printf("%s: received have message\n", c.Peer.String())
		index := binary.BigEndian.Uint32(msg.Payload)
		c.Bitfield.SetPiece(int(index))
		// case message.MsgPiece:
		// 	fmt.Printf("%s: received piece message\n", c.Peer.String())
	}
	return msg, nil
}

// request: <len=0013><id=6><index><begin><length>
func (c *Client) SendRequest(index, begin, length int) error {
	p := make([]byte, 12)
	binary.BigEndian.PutUint32(p[0:4], uint32(index))
	binary.BigEndian.PutUint32(p[4:8], uint32(begin))
	binary.BigEndian.PutUint32(p[8:12], uint32(length))
	m := message.Message{
		ID:      message.MsgRequest,
		Payload: p,
	}
	_, err := c.Conn.Write(m.Serialize())
	return err
}

func (c *Client) Connect() (io.ReadWriteCloser, error) {
	conn, err := net.DialTimeout("tcp", c.Peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func doHandshake(rw io.ReadWriter, h Handshake, p Peer) (Handshake, error) {
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
	if hr.InfoHash != h.InfoHash {
		return Handshake{}, fmt.Errorf("%s: infohash from peer (%s) doesn't match what we asked for (%s)\n",
			p.String(), hr.InfoHash, h.InfoHash)
	}
	return hr, nil
}
