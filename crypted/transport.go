package crypted

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

func SendEncrypted(aead cipher.AEAD, plain []byte) ([] byte, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generování nonce selhalo: %w", err)
	}

	encrypted := aead.Seal(nonce, nonce, plain, nil)

	return encrypted,nil
}

func RecvDecrypted(aead cipher.AEAD, conn *net.UDPConn, buf []byte) ([]byte, error) {
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	if n < aead.NonceSize() {
		return nil, fmt.Errorf("packet příliš krátký: %d bytů", n)
	}

	packet := make([]byte, n)
	copy(packet, buf[:n])

	nonce := packet[:aead.NonceSize()]
	ciphertext := packet[aead.NonceSize():]

	plain, err := aead.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return nil, fmt.Errorf("dešifrování selhalo: %w", err) // autentizační tag selhal → možný útok
	}

	return plain, nil
}