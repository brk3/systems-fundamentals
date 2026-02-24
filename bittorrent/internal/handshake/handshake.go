package handshake

// Handshake is a special message that a peer uses to identify itself
type Handshake struct {
	Pstr     string // protocol identifier
	InfoHash [20]byte
	PeerID   [20]byte
}

// Serialize serializes the handshake to a buffer
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.InfoHash[:])
	copy(buf[curr:], h.PeerID[:])
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
