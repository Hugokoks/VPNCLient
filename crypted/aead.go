package crypted

import (
	"crypto/cipher"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

func NewAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("špatná délka klíče: očekáváno %d, má %d", chacha20poly1305.KeySize, len(key))
	}
	return chacha20poly1305.New(key)
}
