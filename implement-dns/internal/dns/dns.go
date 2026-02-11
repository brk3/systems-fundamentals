package dns

import (
	"bytes"
	"encoding/binary"
)

type DNSHeader struct {
	ID             uint16
	Flags          uint16
	NumQuestions   uint16
	NumAnswers     uint16
	NumAuthorities uint16
	NumAdditionals uint16
}

type DNSQuestion struct {
	Name  []byte
	Type  uint16
	Class uint16
}

func HeaderToBytes(h DNSHeader) ([]byte, error) {
	buf := &bytes.Buffer{}
	// this works because the struct has no slices/strings
	err := binary.Write(buf, binary.BigEndian, h)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func QuestionToBytes(q DNSQuestion) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Write(q.Name) // write the pre-encoded name bytes

	// write the rest as Big Endian
	binary.Write(buf, binary.BigEndian, q.Type)
	binary.Write(buf, binary.BigEndian, q.Class)

	return buf.Bytes(), nil
}

func EncodeDNSName(domainName []byte) []byte {
	buf := &bytes.Buffer{}
	for _, part := range bytes.Split(domainName, []byte(".")) {
		buf.WriteByte(byte(len(part)))
		buf.Write(part)
	}
	buf.WriteByte(0)
	return buf.Bytes()
}
