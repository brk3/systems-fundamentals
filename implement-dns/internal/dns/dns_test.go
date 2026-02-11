package dns

import (
	"bytes"
	"testing"
)

func TestEncodeDNSName(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{
			input:    "google.com",
			expected: []byte{0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00},
		},
		{
			input:    "dns.google.com",
			expected: []byte{0x03, 'd', 'n', 's', 0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00},
		},
		{
			input:    "localhost",
			expected: []byte{0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0x00},
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := EncodeDNSName([]byte(tc.input))
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("For %s, expected %x, got %x", tc.input, tc.expected, result)
			}
		})
	}
}
