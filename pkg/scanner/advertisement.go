package scanner

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"

	"github.com/lazy-electron-consulting/victron-bluetooth/internal"
)

type Advertisement struct {
	Mode             byte
	Model            uint16
	iv               []byte
	keyPrefix        byte
	encryptedPayload []byte
}

// Decrypt decrypts the advert payload using the key.
func (a Advertisement) Decrypt(key []byte) ([]byte, error) {
	if key[0] != a.keyPrefix {
		return nil, fmt.Errorf("key mismatch, expected %x got %x", a.keyPrefix, key[0])
	}
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create aes: %w", err)
	}
	s := cipher.NewOFB(b, internal.AppendZero(a.iv, 16))
	encrypted := internal.AppendZero(a.encryptedPayload, 16)
	decrypted := make([]byte, len(encrypted))
	s.XORKeyStream(decrypted, encrypted)
	return decrypted, nil
}

func readAdvertisement(data []byte) (Advertisement, error) {

	type layout struct {
		_         byte // header
		Mode      byte
		Model     uint16
		Readout   uint8
		IV        [2]byte
		KeyPrefix byte
	}

	l, err := internal.Decode[layout](data)
	if err != nil {
		return Advertisement{}, fmt.Errorf("unable to decode: %w", err)
	}

	return Advertisement{
		Mode:             l.Mode,
		Model:            l.Model,
		iv:               l.IV[:],
		keyPrefix:        l.KeyPrefix,
		encryptedPayload: data[binary.Size(l):],
	}, nil
}
