package bitfield

// A Bitfield represents the pieces that a peer has
type Bitfield []byte

// HasPiece tells if a bitfield has a particular index set
func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	bitIndex := index % 8
	check := bf[byteIndex] >> (7 - bitIndex) // shift target bit all the way to the right
	return check&1 == 1                      // AND with mask of 1 to check if set
}

// SetPiece sets a bit in the bitfield
func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	bitIndex := index % 8
	mask := byte(1 << (7 - bitIndex)) // take a binary 1 and shift the right most bit into position
	bf[byteIndex] |= mask
}
