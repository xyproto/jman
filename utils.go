package jman

import (
	"bytes"
	"strings"
)

// Return the last part of a given JSON path
func lastpart(JSONpath string) string {
	if !strings.Contains(JSONpath, ".") {
		return JSONpath
	}
	parts := strings.Split(JSONpath, ".")
	return parts[len(parts)-1]
}

// Add two byte slices together
func badd(a, b []byte) []byte {
	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)
	return buf.Bytes()
}
