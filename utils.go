package jman

import (
	"bytes"
)

// Add two byte slices together
func badd(a, b []byte) []byte {
	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)
	return buf.Bytes()
}
