package internal

import (
	"bytes"
	"encoding/binary"
)

func Decode[T any](b []byte) (ret T, err error) {
	r := bytes.NewReader(b)
	err = binary.Read(r, binary.LittleEndian, &ret)
	return ret, err
}
