package message

import (
	"encoding/binary"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	mLen := 1 + len(m.Payload)
	buf := make([]byte, 4+mLen) // +4 for length prefix
	binary.BigEndian.PutUint32(buf, uint32(mLen))
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// Read parses a message from a stream. Returns `nil` on keep-alive message
func ReadMessage(r io.Reader) (*Message, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	mLen := binary.BigEndian.Uint32(buf)
	if mLen == 0 {
		return nil, nil
	}
	buf = make([]byte, mLen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return &Message{ID: messageID(buf[0]), Payload: buf[1:]}, nil
}

// piece: <len=0009+X><id=7><index><begin><block>, where X is the length of the block
func ParsePiece(buf []byte, m *Message) int {
	begin := binary.BigEndian.Uint32(m.Payload[4:8])
	block := m.Payload[8:]
	n := copy(buf[begin:], block)
	return n
}
