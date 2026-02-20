package client

import (
	"fmt"
	"io"
	"net"
	"time"

	"go-bt-learning.brk3.github.io/internal/bitfield"
	"go-bt-learning.brk3.github.io/internal/handshake"
	"go-bt-learning.brk3.github.io/internal/message"
	"go-bt-learning.brk3.github.io/internal/peer"
)

// PeerID is a 20 char identifier for our client
const PeerID = "paulsbittorentclient"

type Client struct {
	Conn     io.ReadWriteCloser
	Choked   bool
	Bitfield bitfield.Bitfield
	Peer     peer.Peer
}

func NewClient(peer peer.Peer, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	p := [20]byte{}
	copy(p[:], PeerID)
	h := handshake.Handshake{
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

func (client *Client) HandleMessage() error {
	msg, err := message.ReadMessage(client.Conn)
	if err != nil {
		return err
	}
	switch msg.ID {
	case message.MsgBitfield:
		client.Bitfield = msg.Payload
	case message.MsgUnchoke:
		client.Choked = false
	case message.MsgChoke:
		client.Choked = true
	case message.MsgHave:
		// index, err := message.ParseHave(msg)
		// if err != nil {
		// 	return err
		// }
		// client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		// TODO
		// n, err := message.ParsePiece(state.index, state.buf, msg)
		// state.downloaded += n
		// state.backlog--
	}
	return nil
}

// TODO
func (client *Client) SendRequest(index, begin, length int) error {
	return nil
}

func (c *Client) Connect() (io.ReadWriteCloser, error) {
	conn, err := net.DialTimeout("tcp", c.Peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func doHandshake(rw io.ReadWriter, h handshake.Handshake, p peer.Peer) (handshake.Handshake, error) {
	_, err := rw.Write(h.Serialize())
	if err != nil {
		return handshake.Handshake{}, err
	}
	res := make([]byte, 68)
	_, err = io.ReadFull(rw, res)
	if err != nil {
		return handshake.Handshake{}, err
	}
	hr := handshake.Handshake{}
	hr.Deserialize(res)
	if hr.InfoHash != h.InfoHash {
		return handshake.Handshake{}, fmt.Errorf("%s: infohash from peer (%s) doesn't match what we asked for (%s)\n",
			p.String(), hr.InfoHash, h.InfoHash)
	}
	return hr, nil
}
