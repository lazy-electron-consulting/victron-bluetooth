package internal

import "bytes"

// AppendZero appends 0 bytes to buf until it's length is a multiple of
// blockSize.
func AppendZero(buf []byte, blockSize uint) []byte {
	l := padLen(buf, blockSize)
	if l == 0 {
		return buf
	}
	padding := bytes.Repeat([]byte{0}, int(l))
	return append(buf, padding...)
}

// PrependZero prepends 0 bytes to buf until it's length is a multiple of
// blockSize.
func PrependZero(buf []byte, blockSize uint) []byte {
	l := padLen(buf, blockSize)
	if l == 0 {
		return buf
	}
	padding := bytes.Repeat([]byte{0}, int(l))
	return append(padding, buf...)
}

func padLen(buf []byte, blockSize uint) uint {
	bufLen := uint(len(buf))
	return blockSize - (bufLen % blockSize)
}
