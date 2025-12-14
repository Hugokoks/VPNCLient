package vna

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/chacha20poly1305"
)


func newAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("špatná délka klíče: očekáváno %d, má %d", chacha20poly1305.KeySize, len(key))
	}
	return chacha20poly1305.New(key)
}

func sendEncrypted(aead cipher.AEAD, conn *net.UDPConn, plain []byte) error {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generování nonce selhalo: %w", err)
	}

	out := aead.Seal(nonce, nonce, plain, nil) 

	_, err := conn.Write(out)
	return err
}

func recvDecrypted(aead cipher.AEAD, conn *net.UDPConn, buf []byte) ([]byte, error) {
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	if n < aead.NonceSize() {
		return nil, fmt.Errorf("packet příliš krátký: %d bytů", n)
	}

    packet := make([]byte, n)
    copy(packet,buf[:n])

	nonce := packet[:aead.NonceSize()]
	ciphertext := packet[aead.NonceSize():]

	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	
    if err != nil {
		return nil, fmt.Errorf("dešifrování selhalo: %w", err) // autentizační tag selhal → možný útok
	}

	return plain, nil
}