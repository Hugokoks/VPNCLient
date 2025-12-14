// handshake.go
package vna

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
)


func (v *VNA) Handshake() error {
	if len(v.Keys.ServerPub) == 0 {
		return fmt.Errorf("serverový veřejný klíč není načten")
	}

	//v.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 1. Generuj efemerní privátní klíč klienta
	var clientPriv [32]byte
	if _, err := io.ReadFull(rand.Reader, clientPriv[:]); err != nil {
		return fmt.Errorf("generování efemerního klíče: %w", err)
	}

	// Clamping
	clientPriv[0] &= 248
	clientPriv[31] &= 127
	clientPriv[31] |= 64

	// 2. Vypočítej efemerní public klíč
	var clientPub [32]byte
	curve25519.ScalarBaseMult(&clientPub, &clientPriv)

	// 3. Pošli klientův efemerní pub serveru
	if _, err := v.Conn.Write(clientPub[:]); err != nil {
		return fmt.Errorf("selhalo odeslání client pub: %w", err)
	}

	// 4. Přijmi odpověď serveru: [32]serverEphPub + [64]signature
	response := make([]byte, 96)
	if _, err := io.ReadFull(v.Conn, response); err != nil {
		return fmt.Errorf("selhalo čtení odpovědi serveru: %w", err)
	}

	serverEphPub := response[:32]
	signature := response[32:]

	// 5. Ověř podpis serverem
	if !ed25519.Verify(v.Keys.ServerPub, serverEphPub, signature) {
		return fmt.Errorf("NEPLATNÝ PODPIS SERVERU – možné MITM!")
	}

	// 6. ECDH – spočítej shared secret
	var serverPub [32]byte
	copy(serverPub[:], serverEphPub)
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, &clientPriv, &serverPub)

	// 7. Odvoď symetrický klíč
	h := sha256.Sum256(sharedSecret[:])
	v.Keys.SharedKey = h[:]
	fmt.Printf("Shared key: %x", v.Keys.SharedKey)
	aead, err := newAEAD(v.Keys.SharedKey)
	if err != nil {
	    return err
	}
	v.Aead = aead
	

	fmt.Println("Handshake úspěšný – šifrovací klíč připraven")
	close(v.HandShakeDone)
	return nil
}