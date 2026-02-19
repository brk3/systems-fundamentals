package bittorrent

import "io"

type Client struct {
	Conn     io.ReadWriteCloser
	Choked   bool
	Bitfield Bitfield
	Peer     Peer
}

func NewClient(peer Peer, infoHash [20]byte) (*Client, error) {
	conn, err := peer.Connect()
	if err != nil {
		return nil, err
	}
	h := NewHandshake(infoHash)
	_, err = peer.Handshake(conn, h)
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
	msg, err := ReadMessage(client.Conn)
	if err != nil {
		return err
	}
	switch msg.ID {
	case MsgBitfield:
		client.Bitfield = msg.Payload
	case MsgUnchoke:
		client.Choked = false
	case MsgChoke:
		client.Choked = true
	case MsgHave:
		// TODO
		// index, err := message.ParseHave(msg)
		// state.client.Bitfield.SetPiece(index)
	case MsgPiece:
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
